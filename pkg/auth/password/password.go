package password

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"strconv"

	"github.com/gorilla/securecookie"
	"github.com/sethvargo/go-password/password"
	"golang.org/x/crypto/argon2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
)

func VerifyPassword(ctx context.Context, c client.Client, username string, pass []byte) (verified bool, isDefault bool, err error) {
	secret := corev1.Secret{}
	key := types.NamespacedName{
		Namespace: cosmov1alpha1.UserNamespace(username),
		Name:      cosmov1alpha1.UserPasswordSecretName,
	}

	if err := c.Get(ctx, key, &secret); err != nil {
		return false, isDefault, fmt.Errorf("failed to get password secret: %w", err)
	}

	storedPass, ok := secret.Data[cosmov1alpha1.UserPasswordSecretDataKeyUserPasswordSecret]
	if !ok {
		return false, isDefault, fmt.Errorf("password not found")
	}

	// check is default from annotation
	if ann := secret.GetAnnotations(); ann != nil {
		if val, ok := ann[cosmov1alpha1.UserPasswordSecretAnnKeyUserPasswordIfDefault]; ok {
			if parsedVal, err := strconv.ParseBool(val); err == nil {
				isDefault = parsedVal
			}
		}
	}

	if isDefault {
		return bytesEqual(pass, storedPass), isDefault, nil

	} else {
		salt, ok := secret.Data[cosmov1alpha1.UserPasswordSecretDataKeyUserPasswordSalt]
		if !ok {
			return false, isDefault, fmt.Errorf("salt not found")
		}
		hashedPass, _ := hash(pass, salt)
		return bytesEqual(hashedPass, storedPass), isDefault, nil
	}
}

func bytesEqual(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}

func IsDefaultPassword(ctx context.Context, c client.Client, username string) (bool, error) {
	secret := corev1.Secret{}
	key := types.NamespacedName{
		Namespace: cosmov1alpha1.UserNamespace(username),
		Name:      cosmov1alpha1.UserPasswordSecretName,
	}

	if err := c.Get(ctx, key, &secret); err != nil {
		return false, fmt.Errorf("failed to get password secret: %w", err)
	}

	if ann := secret.GetAnnotations(); ann != nil {
		if val, ok := ann[cosmov1alpha1.UserPasswordSecretAnnKeyUserPasswordIfDefault]; ok {
			if parsedVal, err := strconv.ParseBool(val); err == nil {
				return parsedVal, nil
			}
		}
	}
	return false, nil
}

var ErrNotDefaultPassword = errors.New("not default password")

func GetDefaultPassword(ctx context.Context, c client.Client, username string) (*string, error) {
	secret := corev1.Secret{}
	key := types.NamespacedName{
		Namespace: cosmov1alpha1.UserNamespace(username),
		Name:      cosmov1alpha1.UserPasswordSecretName,
	}

	if err := c.Get(ctx, key, &secret); err != nil {
		return nil, fmt.Errorf("failed to get password secret: %w", err)
	}

	pass, ok := secret.Data[cosmov1alpha1.UserPasswordSecretDataKeyUserPasswordSecret]
	if !ok {
		return nil, fmt.Errorf("password not found")
	}

	var isDefault bool
	if ann := secret.GetAnnotations(); ann != nil {
		if val, ok := ann[cosmov1alpha1.UserPasswordSecretAnnKeyUserPasswordIfDefault]; ok {
			if parsedVal, err := strconv.ParseBool(val); err == nil {
				isDefault = parsedVal
			}
		}
	}

	if !isDefault {
		return nil, ErrNotDefaultPassword
	}

	p := string(pass)
	return &p, nil
}

func ResetPassword(ctx context.Context, c client.Client, username string) error {
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

	return registerPassword(ctx, c, username, []byte(pass), nil)
}

func RegisterPassword(ctx context.Context, c client.Client, username string, password []byte) error {
	hash, salt := hash(password, nil)
	return registerPassword(ctx, c, username, hash, salt)
}

func registerPassword(ctx context.Context, c client.Client, username string, password, salt []byte) error {
	isDefault := false
	if salt == nil {
		isDefault = true
	}

	secret := corev1.Secret{}
	secret.SetName(cosmov1alpha1.UserPasswordSecretName)
	secret.SetNamespace(cosmov1alpha1.UserNamespace(username))

	_, err := ctrl.CreateOrUpdate(ctx, c, &secret, func() error {
		ann := secret.GetAnnotations()
		if ann == nil {
			ann = make(map[string]string)
		}
		ann[cosmov1alpha1.UserPasswordSecretAnnKeyUserPasswordIfDefault] = strconv.FormatBool(isDefault)
		secret.SetAnnotations(ann)

		cosmov1alpha1.SetControllerManaged(&secret)

		secret.Data = map[string][]byte{
			cosmov1alpha1.UserPasswordSecretDataKeyUserPasswordSecret: password,
			cosmov1alpha1.UserPasswordSecretDataKeyUserPasswordSalt:   salt,
		}
		return nil
	})
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
