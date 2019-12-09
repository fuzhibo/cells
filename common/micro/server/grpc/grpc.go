package grpc

// type Server interface {
// 	Options() Options
// 	Init(...Option) error
// 	Handle(Handler) error
// 	NewHandler(interface{}, ...HandlerOption) Handler
// 	NewSubscriber(string, interface{}, ...SubscriberOption) Subscriber
// 	Subscribe(Subscriber) error
// 	Register() error
// 	Deregister() error
// 	Start() error
// 	Stop() error
// 	String() string
// }
import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"time"

	"github.com/micro/util/go/lib/addr"
	"google.golang.org/grpc"

	registry "github.com/micro/go-micro/registry"
	server "github.com/micro/go-micro/server"
)

type grpcServer struct {
	s        *grpc.Server
	opts     server.Options
	handlers map[string]reflect.Type
}

func NewServer(opt ...server.Option) server.Server {
	opts := new(server.Options)
	opts.Metadata = make(map[string]string)
	for _, o := range opt {
		o(opts)
	}

	return &grpcServer{
		s:    grpc.NewServer(),
		opts: *opts,
	}
}

func (s *grpcServer) Options() server.Options {
	return s.opts
}

func (s *grpcServer) Init(...server.Option) error {
	return nil
}

func (s *grpcServer) Handle(h server.Handler) error {

	hdlr := h.Handler()

	// Some micro proto are wrapped so we're bypassing that
	child := reflect.ValueOf(hdlr).Elem().Field(0)
	if child.CanInterface() {
		hdlr = child.Interface()
	}

	s.s.RegisterService(h.(*rpcHandler).getServiceDesc(), hdlr)

	return nil
}

func (s *grpcServer) NewHandler(i interface{}, opts ...server.HandlerOption) server.Handler {
	return newRpcHandler(s.opts.Name, i, opts...)
}

func (s *grpcServer) NewSubscriber(string, interface{}, ...server.SubscriberOption) server.Subscriber {
	return nil
}

func (s *grpcServer) Subscribe(server.Subscriber) error {
	return nil
}

func (s *grpcServer) Register() error {
	hostStr, portStr, err := net.SplitHostPort(s.opts.Address)
	if err != nil {
		return err
	}

	host, _ := addr.Extract(fmt.Sprintf("[%s]", hostStr))
	port, _ := strconv.Atoi(portStr)

	// register service
	node := &registry.Node{
		Id:       s.opts.Name + "-" + s.opts.Id,
		Address:  host,
		Port:     port,
		Metadata: s.opts.Metadata,
	}

	node.Metadata["broker"] = s.opts.Broker.String()
	node.Metadata["registry"] = s.opts.Registry.String()
	node.Metadata["server"] = s.String()
	node.Metadata["transport"] = s.opts.Transport.String()

	service := &registry.Service{
		Name:      s.opts.Name,
		Version:   s.opts.Version,
		Nodes:     []*registry.Node{node},
		Endpoints: []*registry.Endpoint{},
	}

	if err := s.opts.Registry.Register(service); err != nil {
		return err
	}

	return nil
}

func (s *grpcServer) Deregister() error {
	return nil
}

func (s *grpcServer) Start() error {
	lis, err := net.Listen("tcp", s.opts.Address)
	if err != nil {
		return err
	}

	s.opts.Address = lis.Addr().String()

	go func() {
		<-time.After(20 * time.Second)
		s.s.Serve(lis)
	}()
	return nil
}

func (s *grpcServer) Stop() error {
	return nil
}

func (s *grpcServer) String() string {
	return "grpcServer"
}
