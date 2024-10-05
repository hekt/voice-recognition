package util

import "fmt"

func ResourceParent(projectID string) string {
	return fmt.Sprintf("projects/%s/locations/global", projectID)
}

func RecognizerFullname(projectID, recognizerName string) string {
	return fmt.Sprintf("%s/recognizers/%s", ResourceParent(projectID), recognizerName)
}

func PhraseSetFullname(projectID, phraseSetName string) string {
	return fmt.Sprintf("%s/phraseSets/%s", ResourceParent(projectID), phraseSetName)
}
