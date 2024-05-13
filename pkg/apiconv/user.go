package apiconv

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"google.golang.org/protobuf/types/known/timestamppb"
	eventsv1 "k8s.io/api/events/v1"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

type UserConvertOptions func(c *cosmov1alpha1.User, d *dashv1alpha1.User)

func WithUserRaw(withRaw *bool) func(c *cosmov1alpha1.User, d *dashv1alpha1.User) {
	return func(c *cosmov1alpha1.User, d *dashv1alpha1.User) {
		if withRaw != nil && *withRaw {
			d.Raw = ToYAML(c)
		}
	}
}

func C2D_Users(users []cosmov1alpha1.User, opts ...UserConvertOptions) []*dashv1alpha1.User {
	ts := make([]*dashv1alpha1.User, len(users))
	for i, v := range users {
		ts[i] = C2D_User(v, opts...)
	}
	return ts
}

func C2D_User(user cosmov1alpha1.User, opts ...UserConvertOptions) *dashv1alpha1.User {
	d := &dashv1alpha1.User{
		Name:        user.Name,
		DisplayName: user.Spec.DisplayName,
		Roles:       C2S_UserRole(user.Spec.Roles),
		AuthType:    user.Spec.AuthType.String(),
		Addons:      C2D_UserAddons(user.Spec.Addons),
		Status:      string(user.Status.Phase),
	}
	for _, opt := range opts {
		opt(&user, d)
	}
	return d
}

func C2S_UserRole(apiRoles []cosmov1alpha1.UserRole) []string {
	roles := make([]string, 0, len(apiRoles))
	for _, v := range apiRoles {
		roles = append(roles, v.Name)
	}
	return roles
}

func S2C_UserRoles(roles []string) []cosmov1alpha1.UserRole {
	apiRoles := make([]cosmov1alpha1.UserRole, 0, len(roles))
	for _, v := range roles {
		apiRoles = append(apiRoles, cosmov1alpha1.UserRole{Name: v})
	}
	return apiRoles
}

func D2C_UserAddons(addons []*dashv1alpha1.UserAddon) []cosmov1alpha1.UserAddon {
	a := make([]cosmov1alpha1.UserAddon, len(addons))
	for i, v := range addons {
		addon := cosmov1alpha1.UserAddon{
			Template: cosmov1alpha1.UserAddonTemplateRef{
				Name:          v.Template,
				ClusterScoped: v.ClusterScoped,
			},
			Vars: v.Vars,
		}
		a[i] = addon
	}
	return a
}

func C2D_UserAddons(addons []cosmov1alpha1.UserAddon) []*dashv1alpha1.UserAddon {
	da := make([]*dashv1alpha1.UserAddon, len(addons))
	for i, v := range addons {
		da[i] = &dashv1alpha1.UserAddon{
			Template:      v.Template.Name,
			ClusterScoped: v.Template.ClusterScoped,
			Vars:          v.Vars,
		}
	}
	return da
}

func S2D_UserAddons(addons []string) ([]*dashv1alpha1.UserAddon, error) {
	// format
	//   TEMPLATE_NAME
	//   TEMPLATE_NAME,KEY1=XXX,KEY2="YYY ZZZ"
	r1 := regexp.MustCompile(`\w(:\w+=\w+(,\w+=\w+)*)?`)
	r2 := regexp.MustCompile(`^([^= ,]+)=([^,]*)$`)

	userAddons := make([]*dashv1alpha1.UserAddon, 0, len(addons))

	for _, addonParm := range addons {
		if !r1.MatchString(addonParm) {
			return nil, fmt.Errorf("invalid addon format: %s", addonParm)
		}

		addonAndVars := strings.Split(addonParm, ":")
		if addonAndVars[0] == "" {
			return nil, fmt.Errorf("invalid addon format: %s", addonParm)
		}

		userAddon := &dashv1alpha1.UserAddon{
			Template: addonAndVars[0],
		}

		if len(addonAndVars) > 1 {
			addonSplits := strings.Split(addonAndVars[1], ",")
			userAddon.Vars = make(map[string]string, len(addonSplits))

			for _, k_v := range addonSplits {
				kv := r2.FindStringSubmatch(k_v)
				if len(kv) != 3 {
					return nil, fmt.Errorf("invalid addon vars format: %s", k_v)
				}
				_, ok := userAddon.Vars[kv[1]]
				if ok {
					return nil, fmt.Errorf("duplicate addon vars: %s", kv[1])
				}
				userAddon.Vars[kv[1]] = kv[2]
			}
		}
		userAddons = append(userAddons, userAddon)
	}
	return userAddons, nil
}

func D2S_UserAddons(addons []*dashv1alpha1.UserAddon) []string {
	s := make([]string, len(addons))
	for i, addon := range addons {
		t := addon.Template
		kv := make([]string, 0, len(addon.Vars))
		for k, v := range addon.Vars {
			kv = append(kv, fmt.Sprintf("%s=%s", k, v))
		}
		sort.Strings(kv)
		if len(kv) > 0 {
			t = fmt.Sprintf("%s:%s", t, strings.Join(kv, ","))
		}
		s[i] = t
	}
	return s
}

func K2D_Events(events []eventsv1.Event) []*dashv1alpha1.Event {
	es := make([]*dashv1alpha1.Event, len(events))
	for i, v := range events {
		var eventTime *timestamppb.Timestamp
		if v.EventTime.Year() != 1 {
			eventTime = timestamppb.New(v.EventTime.Time)
		} else {
			eventTime = timestamppb.New(v.DeprecatedLastTimestamp.Time)
		}

		var count int32
		if v.Series != nil {
			count = v.Series.Count
		} else {
			count = v.DeprecatedCount
		}

		var lastTime *timestamppb.Timestamp
		if v.Series != nil {
			lastTime = timestamppb.New(v.Series.LastObservedTime.Time)
		} else {
			lastTime = timestamppb.New(v.DeprecatedLastTimestamp.Time)
		}

		e := &dashv1alpha1.Event{
			EventTime: eventTime,
			Reason:    v.Reason,
			Note:      v.Note,
			Type:      v.Type,
			Regarding: &dashv1alpha1.ObjectReference{
				ApiVersion: v.Regarding.APIVersion,
				Kind:       v.Regarding.Kind,
				Name:       v.Regarding.Name,
				Namespace:  v.Regarding.Namespace,
			},
			ReportingController: v.ReportingController,
		}
		if count > 1 {
			e.Series = &dashv1alpha1.EventSeries{
				Count:            count,
				LastObservedTime: lastTime,
			}
		}
		es[i] = e
	}
	return es
}
