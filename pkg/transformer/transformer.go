package transformer

import (
	"context"
	"fmt"

	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Transformer is interface to modify unstructured object
type Transformer interface {
	Transform(*unstructured.Unstructured) (*unstructured.Unstructured, error)
}

// ApplyTransformers applies all transformer to each unstructured objects
func ApplyTransformers(ctx context.Context, transformers []Transformer, objects []unstructured.Unstructured) ([]unstructured.Unstructured, error) {
	log := clog.FromContext(ctx).WithCaller()

	applied := make([]unstructured.Unstructured, len(objects))
	copy(applied, objects)

	for i := 0; i < len(applied); i++ {
		// Perform each transformers
		for _, trans := range transformers {
			transName := Name(trans)
			log.DebugAll().Info(fmt.Sprintf("transforming %s", transName), "transformer", transName, "kind", applied[i].GetKind(), "name", applied[i].GetName())
			before := applied[i].DeepCopy()

			transformed, err := trans.Transform(&applied[i])
			if err != nil {
				return nil, fmt.Errorf("failed to transform: %w", err)
			}

			if !equality.Semantic.DeepEqual(before, transformed) {
				log.DebugAll().PrintObjectDiff(before, transformed)
				log.DebugAll().Info("transformed", "transformer", transName, "kind", applied[i].GetKind(), "name", applied[i].GetName())

				applied[i] = *transformed
			} else {
				log.DebugAll().Info("not transformed", "transformer", transName, "kind", applied[i].GetKind(), "name", applied[i].GetName())
			}
		}
	}
	return applied, nil
}
