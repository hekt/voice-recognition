package util

import "fmt"

func RecognizerParent(projectID string) string {
	return fmt.Sprintf("projects/%s/locations/global", projectID)
}

func RecognizerFullname(projectID, recognizerName string) string {
	return fmt.Sprintf("%s/recognizers/%s", RecognizerParent(projectID), recognizerName)
}

func PhraseSetFullname(projectID, phraseSetName string) string {
	return fmt.Sprintf("%s/phraseSets/%s", RecognizerParent(projectID), phraseSetName)
}
