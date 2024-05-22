package kosmo

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/tools/record"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	"github.com/cosmo-workspace/cosmo/pkg/template"
)

var _ = Describe("event", func() {
	Context("when list user events", func() {
		It("should return a array of the events in user namespace and the user annotated events in default namespace", func() {
			ctx := context.Background()
			err := k8sClient.Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "cosmo-user-xxx"}})
			Expect(err).ShouldNot(HaveOccurred())

			in, err := os.ReadFile("./test/events.yaml")
			Expect(err).ShouldNot(HaveOccurred())

			unst, err := template.NewRawYAMLBuilder(string(in)).Build()
			Expect(err).ShouldNot(HaveOccurred())

			for _, event := range unst {
				RemoveDynamicFields(&event)
				_, err := kubeutil.Apply(ctx, k8sClient, &event, "test", false, false)
				Expect(err).ShouldNot(HaveOccurred(), "failed to create event", event)
			}

			k := NewClient(k8sClient)
			events, err := k.ListEvents(ctx, "cosmo-user-xxx")
			Expect(err).ShouldNot(HaveOccurred())

			data := make([][]string, 0, len(events)+1)
			data = append(data, []string{"Event Time", "Type", "Reason", "Regarding", "Reporting Controller", "Note"})
			for _, v := range events {
				data = append(data, []string{eventTime(v).GoString(), v.Type, v.Reason, v.Regarding.Kind + "/" + v.Regarding.Name, v.ReportingController, v.Note})
			}

			out := bytes.Buffer{}
			w := printers.GetNewTabWriter(&out)

			for _, v := range data {
				fmt.Fprintf(w, "%s\n", strings.Join(v, "\t"))
			}

			w.Flush()
			Expect(out.String()).To(MatchSnapShot())
		})
	})
	Context("Eventf", func() {
		It("emit events", func() {
			eventCount := 4

			rec := record.NewFakeRecorder(eventCount)
			inst := cosmov1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "inst1",
					Namespace: "cosmo-user-tom",
				},
			}
			InstanceEventf(rec, &inst, corev1.EventTypeNormal, "TEST1", "Instance: key1=%s", "value1")

			cinst := cosmov1alpha1.ClusterInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cinst1",
				},
			}
			InstanceEventf(rec, &cinst, corev1.EventTypeNormal, "TEST2", "ClusterInstance: key1=%s", "value1")

			cinst2 := cosmov1alpha1.ClusterInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cinst2",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "cosmo.workspace/v1alpha1",
							Kind:       "User",
							Name:       "user1",
						},
					},
				},
			}
			InstanceEventf(rec, &cinst2, corev1.EventTypeWarning, "TEST3", "Addon: key1=%s", "value1")

			user := cosmov1alpha1.User{
				ObjectMeta: metav1.ObjectMeta{
					Name: "user2",
				},
			}
			UserEventf(rec, &user, corev1.EventTypeWarning, "TEST4", "User: key1=%s", "value1")

			i := 0
		L:
			for {
				select {
				case msg, ok := <-rec.Events:
					if ok {
						i++
						Expect(msg).To(MatchSnapShot())
					} else {
						Fail("CLOSED")
					}
				case <-time.After(time.Second * 10):
					Fail("TIMEOUT")
				default:
					break L
				}
			}
			Expect(i).To(Equal(eventCount))
		})
	})
})

func Test_sorted(t *testing.T) {
	type args struct {
		events []eventsv1.Event
	}
	tests := []struct {
		name string
		args args
		want []eventsv1.Event
	}{
		{
			name: "same DeprecatedFirstTimestamp; diff DeprecatedLastTimestamp",
			args: args{
				events: []eventsv1.Event{
					{
						DeprecatedFirstTimestamp: metav1.Date(2024, 5, 26, 15, 43, 16, 0, time.UTC),
						DeprecatedLastTimestamp:  metav1.Date(2024, 5, 26, 15, 44, 35, 0, time.UTC),
						DeprecatedSource: corev1.EventSource{
							Component: "replicaset-controller",
						},
						DeprecatedCount: 6,
						Note:            "(combined from similar events): Error creating: pods \"ws2-workspace-75dc7c5469-5trff\" is forbidden: failed quota: quota: must specify limits.cpu for: code-server",
						Reason:          "FailedCreate",
						Type:            "Warning",
					},
					{
						DeprecatedFirstTimestamp: metav1.Date(2024, 5, 26, 15, 43, 16, 0, time.UTC),
						DeprecatedLastTimestamp:  metav1.Date(2024, 5, 26, 15, 43, 16, 0, time.UTC),
						DeprecatedSource: corev1.EventSource{
							Component: "replicaset-controller",
						},
						DeprecatedCount: 1,
						Note:            "Error creating: pods \"ws2-workspace-75dc7c5469-cxh42\" is forbidden: failed quota: quota: must specify limits.cpu for: code-server",
						Reason:          "FailedCreate",
						Type:            "Warning",
					},
				},
			},
			want: []eventsv1.Event{
				{
					DeprecatedFirstTimestamp: metav1.Date(2024, 5, 26, 15, 43, 16, 0, time.UTC),
					DeprecatedLastTimestamp:  metav1.Date(2024, 5, 26, 15, 43, 16, 0, time.UTC),
					DeprecatedSource: corev1.EventSource{
						Component: "replicaset-controller",
					},
					DeprecatedCount: 1,
					Note:            "Error creating: pods \"ws2-workspace-75dc7c5469-cxh42\" is forbidden: failed quota: quota: must specify limits.cpu for: code-server",
					Reason:          "FailedCreate",
					Type:            "Warning",
				},
				{
					DeprecatedFirstTimestamp: metav1.Date(2024, 5, 26, 15, 43, 16, 0, time.UTC),
					DeprecatedLastTimestamp:  metav1.Date(2024, 5, 26, 15, 44, 35, 0, time.UTC),
					DeprecatedSource: corev1.EventSource{
						Component: "replicaset-controller",
					},
					DeprecatedCount: 6,
					Note:            "(combined from similar events): Error creating: pods \"ws2-workspace-75dc7c5469-5trff\" is forbidden: failed quota: quota: must specify limits.cpu for: code-server",
					Reason:          "FailedCreate",
					Type:            "Warning",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorted(tt.args.events)
			if !slices.EqualFunc(tt.args.events, tt.want, func(a, b eventsv1.Event) bool { return reflect.DeepEqual(a, b) }) {
				t.Errorf("sorted() diff %s", cmp.Diff(tt.args.events, tt.want))
			}
		})
	}
}
