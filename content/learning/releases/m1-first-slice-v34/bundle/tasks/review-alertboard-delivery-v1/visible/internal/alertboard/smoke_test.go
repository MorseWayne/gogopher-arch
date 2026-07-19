package alertboard

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type smokeStore struct{}
func(smokeStore)Alert(context.Context,string,string)(Alert,error){return Alert{},ErrNotFound}
func(smokeStore)Acknowledge(context.Context,string,string)error{return ErrNotFound}
func(smokeStore)Next(ctx context.Context)(Delivery,error){<-ctx.Done();return Delivery{},ctx.Err()}
func(smokeStore)Complete(context.Context,string,error)error{return nil}

func TestUnauthenticatedAlertIsRejected(t *testing.T){service,err:=NewService(smokeStore{},nil,map[string]string{"tenant":"secret"});if err!=nil{t.Fatal(err)};response:=httptest.NewRecorder();service.Handler().ServeHTTP(response,httptest.NewRequest(http.MethodGet,"/v1/alerts/a-1",nil));if response.Code!=http.StatusUnauthorized{t.Fatalf("status=%d",response.Code)}}
