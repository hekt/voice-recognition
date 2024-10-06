package testutil

import (
	"context"
	"net"
	"testing"

	speech "cloud.google.com/go/speech/apiv2"
	"cloud.google.com/go/speech/apiv2/speechpb"
	myspeech "github.com/hekt/voice-recognition/internal/interfaces/speech"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func MockSpeechClient(t *testing.T, ctx context.Context, mockServer speechpb.SpeechServer) myspeech.Client {
	t.Helper()

	l := bufconn.Listen(1024 * 1024)
	t.Cleanup(func() { l.Close() })

	s := grpc.NewServer()
	speechpb.RegisterSpeechServer(s, mockServer)

	go s.Serve(l)
	t.Cleanup(func() { s.Stop() })

	conn, err := grpc.NewClient(
		// use passthrough resolver explicitly to avoid using default dns resolver
		// ref. https://stackoverflow.com/questions/78485578/how-to-use-the-bufconn-package-with-grpc-newclient
		"passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return l.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	client, err := speech.NewClient(ctx, option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	t.Cleanup(func() { client.Close() })

	return client
}

func AnyResponse(t *testing.T, resp proto.Message) *anypb.Any {
	t.Helper()

	anyResp, err := anypb.New(resp)
	if err != nil {
		t.Fatalf("failed to create anypb: %v", err)
	}

	return anyResp
}
