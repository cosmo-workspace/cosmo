package kubeutil

import (
	"context"
	"reflect"
	"regexp"
	"runtime"

	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClientMock struct {
	client.Client

	GetMock         func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) (mocked bool, err error)
	ListMock        func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (mocked bool, err error)
	CreateMock      func(ctx context.Context, obj client.Object, opts ...client.CreateOption) (mocked bool, err error)
	DeleteMock      func(ctx context.Context, obj client.Object, opts ...client.DeleteOption) (mocked bool, err error)
	UpdateMock      func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) (mocked bool, err error)
	PatchMock       func(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) (mocked bool, err error)
	DeleteAllOfMock func(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) (mocked bool, err error)
}

func NewClientMock(c client.Client) ClientMock {
	return ClientMock{Client: c}
}

func (c *ClientMock) Clear() {
	ctx := context.Background()
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("Client mock clear")
	c.GetMock = nil
	c.ListMock = nil
	c.CreateMock = nil
	c.DeleteMock = nil
	c.UpdateMock = nil
	c.PatchMock = nil
	c.DeleteAllOfMock = nil
}

func (c *ClientMock) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if c.GetMock != nil {
		mocked, err := c.GetMock(ctx, key, obj)
		if mocked {
			log := clog.FromContext(ctx).WithCaller()
			log.Debug().Info("GetMock")
			return err
		}
	}
	return c.Client.Get(ctx, key, obj, opts...)
}

func (c *ClientMock) SetGetError(caller interface{}, retErr error) {
	c.GetMock = func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) (mocked bool, err error) {
		if c.IsCallingFrom(caller) {
			return true, retErr
		}
		return false, nil
	}
}

func (c *ClientMock) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if c.ListMock != nil {
		mocked, err := c.ListMock(ctx, list, opts...)
		if mocked {
			log := clog.FromContext(ctx).WithCaller()
			log.Debug().Info("ListMock")
			return err
		}
	}
	return c.Client.List(ctx, list, opts...)
}

func (c *ClientMock) SetListError(caller interface{}, retErr error) {
	c.ListMock = func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (mocked bool, err error) {
		if c.IsCallingFrom(caller) {
			return true, retErr
		}
		return false, nil
	}
}

func (c *ClientMock) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if c.CreateMock != nil {
		mocked, err := c.CreateMock(ctx, obj, opts...)
		if mocked {
			log := clog.FromContext(ctx).WithCaller()
			log.Debug().Info("CreateMock")
			return err
		}
	}
	return c.Client.Create(ctx, obj, opts...)
}

func (c *ClientMock) SetCreateError(caller interface{}, retErr error) {
	c.CreateMock = func(ctx context.Context, obj client.Object, opts ...client.CreateOption) (mocked bool, err error) {
		if c.IsCallingFrom(caller) {
			return true, retErr
		}
		return false, nil
	}
}

func (c *ClientMock) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	if c.DeleteMock != nil {
		mocked, err := c.DeleteMock(ctx, obj, opts...)
		if mocked {
			log := clog.FromContext(ctx).WithCaller()
			log.Debug().Info("DeleteMock")
			return err
		}
	}
	return c.Client.Delete(ctx, obj, opts...)
}

func (c *ClientMock) SetDeleteError(caller interface{}, retErr error) {
	c.DeleteMock = func(ctx context.Context, obj client.Object, opts ...client.DeleteOption) (mocked bool, err error) {
		if c.IsCallingFrom(caller) {
			return true, retErr
		}
		return false, nil
	}
}

func (c *ClientMock) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if c.UpdateMock != nil {
		mocked, err := c.UpdateMock(ctx, obj, opts...)
		if mocked {
			log := clog.FromContext(ctx).WithCaller()
			log.Debug().Info("UpdateMock")
			return err
		}
	}
	return c.Client.Update(ctx, obj, opts...)
}

func (c *ClientMock) SetUpdateError(caller interface{}, retErr error) {
	c.UpdateMock = func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) (mocked bool, err error) {
		if c.IsCallingFrom(caller) {
			return true, retErr
		}
		return false, nil
	}
}

func (c *ClientMock) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	if c.PatchMock != nil {
		mocked, err := c.PatchMock(ctx, obj, patch, opts...)
		if mocked {
			log := clog.FromContext(ctx).WithCaller()
			log.Debug().Info("PatchMock")
			return err
		}
	}
	return c.Client.Patch(ctx, obj, patch, opts...)
}

func (c *ClientMock) SetPatchError(caller interface{}, retErr error) {
	c.PatchMock = func(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) (mocked bool, err error) {
		if c.IsCallingFrom(caller) {
			return true, retErr
		}
		return false, nil
	}
}

func (c *ClientMock) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	if c.DeleteAllOfMock != nil {
		mocked, err := c.DeleteAllOfMock(ctx, obj, opts...)
		if mocked {
			log := clog.FromContext(ctx).WithCaller()
			log.Debug().Info("DeleteAllOfMock")
			return err
		}
	}
	return c.Client.DeleteAllOf(ctx, obj, opts...)
}

func (c *ClientMock) SetDeleteAllOfError(caller interface{}, retErr error) {
	c.DeleteAllOfMock = func(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) (mocked bool, err error) {
		if c.IsCallingFrom(caller) {
			return true, retErr
		}
		return false, nil
	}
}

func (c *ClientMock) IsCallingFrom(f interface{}) bool {
	var re *regexp.Regexp
	switch v := f.(type) {
	case string:
		re = regexp.MustCompile(v)
	default:
		fv := reflect.ValueOf(v).Pointer()
		fvname := runtime.FuncForPC(fv).Name()
		fvname = regexp.QuoteMeta(fvname)
		re = regexp.MustCompile(fvname)
	}

	for i := 1; i < 9999; i++ {
		pc, _, _, ok := runtime.Caller(i)
		if !ok {
			break
		}
		pcname := runtime.FuncForPC(pc).Name()
		if re.MatchString(pcname) {
			return true
		}
	}
	return false
}
