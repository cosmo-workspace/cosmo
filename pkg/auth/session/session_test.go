package session_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/cosmo-workspace/cosmo/pkg/auth/session"
	"github.com/gorilla/sessions"
)

func TestNewStore(t *testing.T) {
	type args struct {
		hashKey  []byte
		blockKey []byte
		opt      *http.Cookie
	}
	tests := []struct {
		name string
		args args
		want sessions.Store
	}{
		{
			name: "✅ ok",
			args: args{
				hashKey:  []byte("1234567890"),
				blockKey: []byte("abcdefghij"),
				opt:      &http.Cookie{Name: "nnn", Domain: "ddd"},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := session.NewStore(tt.args.hashKey, tt.args.blockKey, tt.args.opt)
			sess, _ := store.New(&http.Request{}, "name")
			sesInfo := session.Info{
				UserName: "user1",
				Deadline: time.Date(2999, 4, 1, 0, 0, 0, 0, time.Local).Unix(),
			}
			session.Set(sess, sesInfo)
			store.Save(&http.Request{}, httptest.NewRecorder(), sess)
			sesInfo2 := session.Get(sess)
			if !reflect.DeepEqual(sesInfo, sesInfo2) {
				t.Errorf("NewStore() = %v, want %v", sesInfo, sesInfo2)
			}
		})
	}
}
