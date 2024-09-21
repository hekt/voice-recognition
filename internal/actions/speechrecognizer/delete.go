package speechrecognizer

import (
	"context"
	"fmt"

	speech "cloud.google.com/go/speech/apiv2"
	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/hekt/voice-recognition/internal/util"
)

type DeleteArgs struct {
	ProjectID      string
	RecognizerName string
}

func Delete(ctx context.Context, args DeleteArgs) error {
	client, err := speech.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	op, err := client.DeleteRecognizer(ctx, &speechpb.DeleteRecognizerRequest{
		Name: util.RecognizerFullname(args.ProjectID, args.RecognizerName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete recognizer: %w", err)
	}

	if _, e := op.Wait(ctx); e != nil {
		return fmt.Errorf("failed to wait for delete operation: %w", e)
	}

	fmt.Println("Recognizer deleted")

	return nil
}
