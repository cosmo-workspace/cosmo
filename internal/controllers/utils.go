package controllers

import (
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ignoreNotFound return nil if the given err is NotFoundErr.
func ignoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}

type object interface {
	GetName() string
	GetObjectKind() schema.ObjectKind
}
