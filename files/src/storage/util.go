package storage

func GetPublicFileURL(key string) string {
	return cdnBaseURL + "/" + key
}
