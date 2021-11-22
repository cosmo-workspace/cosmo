package transformer

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Test_ToUnstructured(t *testing.T) {
	type args struct {
		obj interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "OK",
			args: args{
				obj: &corev1.ServicePort{
					Name:       "port",
					Port:       int32(8080),
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(8080),
				},
			},
			want: map[string]interface{}{
				"name":       "port",
				"port":       8080,
				"protocol":   "TCP",
				"targetPort": 8080,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToUnstructured(tt.args.obj)
			if err != nil {
				t.Errorf("ToUnstructured() err = %v", err)
				return
			}
			if reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToUnstructured() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNestedMap(t *testing.T) {
	type args struct {
		objMap map[string]interface{}
		path   string
	}
	tests := []struct {
		name  string
		args  args
		want  map[string]interface{}
		want1 bool
	}{
		{
			name: "found",
			args: args{
				objMap: map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": "foobar",
					},
				},
				path: "foo",
			},
			want: map[string]interface{}{
				"bar": "foobar",
			},
			want1: true,
		},
		{
			name: "nested",
			args: args{
				objMap: map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": map[string]interface{}{
							"foobar": map[string]interface{}{
								"key": "val",
							},
						},
					},
				},
				path: "foo.bar",
			},
			want: map[string]interface{}{
				"foobar": map[string]interface{}{
					"key": "val",
				},
			},
			want1: true,
		},
		{
			name: "not found",
			args: args{
				objMap: map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": "foobar",
					},
				},
				path: "bar",
			},
			want:  nil,
			want1: false,
		},
		{
			name: "not found nested",
			args: args{
				objMap: map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": "foobar",
					},
				},
				path: "foo.foobar",
			},
			want:  nil,
			want1: false,
		},
		{
			name: "not found nested",
			args: args{
				objMap: map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": "foobar",
					},
				},
				path: "foo.bar.foobar",
			},
			want:  nil,
			want1: false,
		},
		{
			name: "not found nested",
			args: args{
				objMap: map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": "foobar",
					},
				},
				path: "bar.foobar",
			},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := NestedMap(tt.args.objMap, tt.args.path)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NestedMap() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("NestedMap() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNestedSlice(t *testing.T) {
	type args struct {
		objMap map[string]interface{}
		path   string
	}
	tests := []struct {
		name  string
		args  args
		want  []interface{}
		want1 bool
	}{
		{
			name: "found",
			args: args{
				objMap: map[string]interface{}{
					"foo": []interface{}{
						"bar1", "bar2", "bar3",
					},
				},
				path: "foo",
			},
			want: []interface{}{
				"bar1", "bar2", "bar3",
			},
			want1: true,
		},
		{
			name: "found nested",
			args: args{
				objMap: map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": map[string]interface{}{
							"foobar": []interface{}{
								"bar1", "bar2", "bar3",
							},
						},
					},
				},
				path: "foo.bar.foobar",
			},
			want: []interface{}{
				"bar1", "bar2", "bar3",
			},
			want1: true,
		},
		{
			name: "not found",
			args: args{
				objMap: map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": map[string]interface{}{
							"foobar": []interface{}{
								"bar1", "bar2", "bar3",
							},
						},
					},
				},
				path: "bar",
			},
			want:  nil,
			want1: false,
		},
		{
			name: "not found nested nokey",
			args: args{
				objMap: map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": map[string]interface{}{
							"foobar": []interface{}{
								"bar1", "bar2", "bar3",
							},
						},
					},
				},
				path: "foo.foobar",
			},
			want:  nil,
			want1: false,
		},
		{
			name: "not found nested not map",
			args: args{
				objMap: map[string]interface{}{
					"foo": map[string]interface{}{
						"bar1": map[string]interface{}{
							"foobar": []interface{}{
								"bar1", "bar2", "bar3",
							},
						},
						"bar2": map[string]interface{}{
							"foobar": []interface{}{
								"bar1", "bar2", "bar3",
							},
						},
					},
				},
				path: "foo.bar.not",
			},
			want:  nil,
			want1: false,
		},
		{
			name: "not found nested not map",
			args: args{
				objMap: map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": map[string]interface{}{
							"foobar": []interface{}{
								"bar1", "bar2", "bar3",
							},
						},
					},
				},
				path: "foo.bar.foobar.not",
			},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := NestedSlice(tt.args.objMap, tt.args.path)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NestedSlice() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("NestedSlice() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNestedMapDelete(t *testing.T) {
	type args struct {
		objMap map[string]interface{}
		path   string
	}
	tests := []struct {
		name     string
		args     args
		wantBool bool
		want     map[string]interface{}
	}{
		{
			name: "delete",
			args: args{
				objMap: map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": map[string]interface{}{
							"foobar": []interface{}{
								"bar1", "bar2", "bar3",
							},
							"foobar2": []interface{}{
								"bar1", "bar2", "bar3",
							},
						},
					},
				},
				path: "foo.bar.foobar",
			},
			want: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{
						"foobar2": []interface{}{
							"bar1", "bar2", "bar3",
						},
					},
				},
			},
			wantBool: true,
		},
		{
			name: "not delete not map",
			args: args{
				objMap: map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": map[string]interface{}{
							"foobar": []interface{}{
								"bar1", "bar2", "bar3",
							},
							"foobar2": []interface{}{
								"bar1", "bar2", "bar3",
							},
						},
					},
				},
				path: "foo.bar.foobar.bar1",
			},
			want: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{
						"foobar": []interface{}{
							"bar1", "bar2", "bar3",
						},
						"foobar2": []interface{}{
							"bar1", "bar2", "bar3",
						},
					},
				},
			},
			wantBool: false,
		},
		{
			name: "not delete not found 2",
			args: args{
				objMap: map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": map[string]interface{}{
							"foobar": []interface{}{
								"bar1", "bar2", "bar3",
							},
							"foobar2": []interface{}{
								"bar1", "bar2", "bar3",
							},
						},
					},
				},
				path: "foo.bar2.foobar",
			},
			want: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{
						"foobar": []interface{}{
							"bar1", "bar2", "bar3",
						},
						"foobar2": []interface{}{
							"bar1", "bar2", "bar3",
						},
					},
				},
			},
			wantBool: false,
		},
		{
			name: "not delete not found 3",
			args: args{
				objMap: map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": map[string]interface{}{
							"foobar": []interface{}{
								"bar1", "bar2", "bar3",
							},
							"foobar2": []interface{}{
								"bar1", "bar2", "bar3",
							},
						},
					},
				},
				path: "foo.bar.foobar3",
			},
			want: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{
						"foobar": []interface{}{
							"bar1", "bar2", "bar3",
						},
						"foobar2": []interface{}{
							"bar1", "bar2", "bar3",
						},
					},
				},
			},
			wantBool: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := tt.args.objMap
			gotBool := NestedMapDelete(mod, tt.args.path)
			if gotBool != tt.wantBool {
				t.Errorf("NestedMapDelete() gotBool = %v, want %v", gotBool, tt.wantBool)
				t.Errorf("NestedMapDelete() got = %v, want %v", mod, tt.want)
			}
			if !reflect.DeepEqual(mod, tt.want) {
				t.Errorf("NestedMapDelete() got = %v, want %v", mod, tt.want)
			}
		})
	}
}

func TestName(t *testing.T) {
	t.Run("ScalingTransformer", func(t *testing.T) {
		tf := NewScalingTransformer(nil, "")
		if got := Name(tf); got != "ScalingTransformer" {
			t.Errorf("Name() = %v, want ScalingTransformer", got)
		}
	})
	t.Run("ok", func(t *testing.T) {
		tf := NewMetadataTransformer(nil, nil, nil)
		if got := Name(tf); got != "MetadataTransformer" {
			t.Errorf("Name() = %v, want MetadataTransformer", got)
		}
	})
}
