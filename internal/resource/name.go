package resource

import "fmt"

func ParentName(projectID string) string {
	return fmt.Sprintf("projects/%s/locations/global", projectID)
}

func RecognizerFullname(projectID, recognizerName string) string {
	return fmt.Sprintf("%s/recognizers/%s", ParentName(projectID), recognizerName)
}

func PhraseSetFullname(projectID, phraseSetName string) string {
	return fmt.Sprintf("%s/phraseSets/%s", ParentName(projectID), phraseSetName)
}
