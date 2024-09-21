package speechrecognizer

import (
	"context"
	"fmt"

	speech "cloud.google.com/go/speech/apiv2"
	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/hekt/voice-recognition/internal/util"
	"google.golang.org/api/iterator"
)

type ListArgs struct {
	ProjectID string
}

func List(ctx context.Context, args ListArgs) error {
	client, err := speech.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	iterResp := client.ListRecognizers(ctx, &speechpb.ListRecognizersRequest{
		Parent:      util.RecognizerParent(args.ProjectID),
		ShowDeleted: true,
	})

	for {
		resp, err := iterResp.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return fmt.Errorf("failed to get next response: %w", err)
		}

		fmt.Printf("Recognizer: %v\n", resp)
	}

	return nil
}
