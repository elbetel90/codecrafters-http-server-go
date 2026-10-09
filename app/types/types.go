package types

type HttpMethod string

const (
	HttpMethodPOST HttpMethod = "POST"
	HttpMethodGET  HttpMethod = "GET"
)

type CompressionScheme string

const (
	CompressionSchemeGzip CompressionScheme = "gzip"
)

func SupportedCompressionSchemes() []string {
	return []string{
		string(CompressionSchemeGzip),
	}
}
