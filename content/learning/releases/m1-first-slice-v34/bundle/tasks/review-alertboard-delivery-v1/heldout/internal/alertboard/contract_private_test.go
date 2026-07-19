package alertboard

import(
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type variantStore struct{mu sync.Mutex;alerts map[string]Alert;lookups,acks int;jobs chan Delivery;completed chan string;ackErr error}
func(store *variantStore)Alert(_ context.Context,tenant,id string)(Alert,error){store.mu.Lock();defer store.mu.Unlock();store.lookups++;alert,ok:=store.alerts[tenant+"/"+id];if !ok{return Alert{},ErrNotFound};return alert,nil}
func(store *variantStore)Acknowledge(_ context.Context,tenant,id string)error{store.mu.Lock();defer store.mu.Unlock();store.acks++;if store.ackErr!=nil{return store.ackErr};alert,ok:=store.alerts[tenant+"/"+id];if !ok{return ErrNotFound};alert.Acknowledged=true;store.alerts[tenant+"/"+id]=alert;return nil}
func(store *variantStore)Next(ctx context.Context)(Delivery,error){select{case job:=<-store.jobs:return job,nil;case<-ctx.Done():return Delivery{},ctx.Err()}}
func(store *variantStore)Complete(_ context.Context,id string,_ error)error{store.completed<-id;return nil}
type variantCache struct{mu sync.Mutex;values map[string]Alert;gets,sets,deletes int;err error}
func(cache *variantCache)Get(_ context.Context,key string)(Alert,bool,error){cache.mu.Lock();defer cache.mu.Unlock();cache.gets++;if cache.err!=nil{return Alert{},false,cache.err};alert,ok:=cache.values[key];return alert,ok,nil}
func(cache *variantCache)Set(_ context.Context,key string,alert Alert,_ time.Duration)error{cache.mu.Lock();defer cache.mu.Unlock();cache.sets++;if cache.values==nil{cache.values=map[string]Alert{}};cache.values[key]=alert;return cache.err}
func(cache *variantCache)Delete(_ context.Context,key string)error{cache.mu.Lock();defer cache.mu.Unlock();cache.deletes++;delete(cache.values,key);return cache.err}
func variantRequest(handler http.Handler,method,path,key string)*httptest.ResponseRecorder{request:=httptest.NewRequest(method,path,nil);if key!=""{request.Header.Set("X-API-Key",key)};response:=httptest.NewRecorder();handler.ServeHTTP(response,request);return response}

func TestVariantServicePreservesSecurityCacheAndWorkerBoundaries(t *testing.T){store:=&variantStore{alerts:map[string]Alert{"one/a-1":{ID:"a-1",TenantID:"one",Message:"disk"}},jobs:make(chan Delivery,4),completed:make(chan string,4)};cache:=&variantCache{values:map[string]Alert{}};service,err:=NewService(store,cache,map[string]string{"one":"secret-one","two":"secret-two"});if err!=nil{t.Fatal(err)};handler:=service.Handler();if got:=variantRequest(handler,http.MethodGet,"/v1/alerts/a-1","");got.Code!=http.StatusUnauthorized||store.lookups!=0{t.Fatalf("unauthorized=%d lookups=%d",got.Code,store.lookups)};if got:=variantRequest(handler,http.MethodGet,"/v1/alerts/a-1","secret-two");got.Code!=http.StatusNotFound{t.Fatalf("cross tenant=%d",got.Code)};first:=variantRequest(handler,http.MethodGet,"/v1/alerts/a-1","secret-one");second:=variantRequest(handler,http.MethodGet,"/v1/alerts/a-1","secret-one");if first.Code!=200||second.Code!=200||store.lookups!=2||cache.sets!=1{t.Fatalf("cache status=%d/%d lookups=%d sets=%d",first.Code,second.Code,store.lookups,cache.sets)};store.ackErr=errors.New("write failed");before:=cache.deletes;if got:=variantRequest(handler,http.MethodPost,"/v1/alerts/a-1/ack","secret-one");got.Code!=http.StatusInternalServerError||cache.deletes!=before{t.Fatalf("failed ack status=%d deletes=%d",got.Code,cache.deletes-before)};store.ackErr=nil;if got:=variantRequest(handler,http.MethodPost,"/v1/alerts/a-1/ack","secret-one");got.Code!=http.StatusNoContent||cache.deletes!=before+1{t.Fatalf("ack status=%d deletes=%d",got.Code,cache.deletes-before)};for index:=0;index<4;index++{store.jobs<-Delivery{ID:string(rune('a'+index))}};ctx,cancel:=context.WithCancel(context.Background());release:=make(chan struct{});var active,max atomic.Int32;done:=make(chan error,1);go func(){done<-service.Run(ctx,2,func(context.Context,Delivery)error{current:=active.Add(1);for{old:=max.Load();if current<=old||max.CompareAndSwap(old,current){break}};<-release;active.Add(-1);return nil})}();deadline:=time.After(time.Second);for max.Load()<2{select{case<-deadline:t.Fatal("workers not started");default:time.Sleep(time.Millisecond)}};if max.Load()>2{t.Fatalf("max=%d",max.Load())};close(release);for index:=0;index<2;index++{<-store.completed};cancel();select{case err:=<-done:if err!=nil{t.Fatal(err)};case<-time.After(time.Second):t.Fatal("workers did not join")}}
