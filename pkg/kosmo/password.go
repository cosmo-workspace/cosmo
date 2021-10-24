package kosmo

import (
	"context"
	"crypto/subtle"
	"fmt"
	"strconv"

	"github.com/gorilla/securecookie"
	"github.com/sethvargo/go-password/password"
	"golang.org/x/crypto/argon2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
)

func (c *Client) VerifyPassword(ctx context.Context, userid string, pass []byte) (verified bool, isDefault bool, err error) {
	secret := corev1.Secret{}
	key := types.NamespacedName{
		Namespace: wsv1alpha1.UserNamespace(userid),
		Name:      wsv1alpha1.UserPasswordSecretName,
	}

	if err := c.Get(ctx, key, &secret); err != nil {
		return false, isDefault, fmt.Errorf("failed to get password secret: %w", err)
	}

	storedPass, ok := secret.Data[wsv1alpha1.UserPasswordSecretDataKeyUserPasswordSecret]
	if !ok {
		return false, isDefault, fmt.Errorf("password not found")
	}

	// check is default from annotation
	if ann := secret.GetAnnotations(); ann != nil {
		if val, ok := ann[wsv1alpha1.UserPasswordSecretAnnKeyUserPasswordIfDefault]; ok {
			if parsedVal, err := strconv.ParseBool(val); err == nil {
				isDefault = parsedVal
			}
		}
	}

	if isDefault {
		return BytesEqual(pass, storedPass), isDefault, nil

	} else {
		salt, ok := secret.Data[wsv1alpha1.UserPasswordSecretDataKeyUserPasswordSalt]
		if !ok {
			return false, isDefault, fmt.Errorf("salt not found")
		}
		hashedPass, _ := hash(pass, salt)
		return BytesEqual(hashedPass, storedPass), isDefault, nil
	}
}

func BytesEqual(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}

func (c *Client) IsDefaultPassword(ctx context.Context, userid string) (bool, error) {
	secret := corev1.Secret{}
	key := types.NamespacedName{
		Namespace: wsv1alpha1.UserNamespace(userid),
		Name:      wsv1alpha1.UserPasswordSecretName,
	}

	if err := c.Get(ctx, key, &secret); err != nil {
		return false, fmt.Errorf("failed to get password secret: %w", err)
	}

	if ann := secret.GetAnnotations(); ann != nil {
		if val, ok := ann[wsv1alpha1.UserPasswordSecretAnnKeyUserPasswordIfDefault]; ok {
			if parsedVal, err := strconv.ParseBool(val); err == nil {
				return parsedVal, nil
			}
		}
	}
	return false, nil
}

func (c *Client) GetDefaultPassword(ctx context.Context, userid string) (*string, error) {
	secret := corev1.Secret{}
	key := types.NamespacedName{
		Namespace: wsv1alpha1.UserNamespace(userid),
		Name:      wsv1alpha1.UserPasswordSecretName,
	}

	if err := c.Get(ctx, key, &secret); err != nil {
		return nil, fmt.Errorf("failed to get password secret: %w", err)
	}

	pass, ok := secret.Data[wsv1alpha1.UserPasswordSecretDataKeyUserPasswordSecret]
	if !ok {
		return nil, fmt.Errorf("password not found")
	}

	var isDefault bool
	if ann := secret.GetAnnotations(); ann != nil {
		if val, ok := ann[wsv1alpha1.UserPasswordSecretAnnKeyUserPasswordIfDefault]; ok {
			if parsedVal, err := strconv.ParseBool(val); err == nil {
				isDefault = parsedVal
			}
		}
	}

	if !isDefault {
		return nil, fmt.Errorf("not default")
	}

	p := string(pass)
	return &p, nil
}

func (c *Client) ResetPassword(ctx context.Context, userid string) error {
	g, err := password.NewGenerator(&password.GeneratorInput{
		Symbols: "~!@#$%^&*()_+-={}|[]:<>?,./",
	})
	if err != nil {
		return fmt.Errorf("failed to create password generator: %w", err)
	}

	pass, err := g.Generate(16, 3, 3, false, false)
	if err != nil {
		return fmt.Errorf("failed to generate random password: %w", err)
	}

	return c.registerPassword(ctx, userid, []byte(pass), nil)
}

func (c *Client) RegisterPassword(ctx context.Context, userid string, password []byte) error {
	hash, salt := hash(password, nil)
	return c.registerPassword(ctx, userid, hash, salt)
}

func (c *Client) registerPassword(ctx context.Context, userid string, password, salt []byte) error {
	log := clog.FromContext(ctx).WithCaller()

	isDefault := false
	if salt == nil {
		isDefault = true
	}

	secret := corev1.Secret{}
	secret.SetName(wsv1alpha1.UserPasswordSecretName)
	secret.SetNamespace(wsv1alpha1.UserNamespace(userid))

	op, err := ctrl.CreateOrUpdate(ctx, c.Client, &secret, func() error {
		secret.Annotations = map[string]string{
			wsv1alpha1.UserPasswordSecretAnnKeyUserPasswordIfDefault: strconv.FormatBool(isDefault)}

		secret.Data = map[string][]byte{
			wsv1alpha1.UserPasswordSecretDataKeyUserPasswordSecret: password,
			wsv1alpha1.UserPasswordSecretDataKeyUserPasswordSalt:   salt,
		}
		return nil
	})
	log.Info("register password secret", "name", secret.Name, "ns", secret.Namespace, "op", op, "error", err)

	return err
}

type argon2params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	keyLength   uint32
	saltLength  int
}

func hash(password, salt []byte) ([]byte, []byte) {
	p := argon2params{
		memory:      2048,
		iterations:  1,
		parallelism: 4,
		saltLength:  128,
		keyLength:   256,
	}

	if salt == nil {
		salt = securecookie.GenerateRandomKey(p.saltLength)
	}

	hash := argon2.IDKey(password, salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	return hash, salt
}
