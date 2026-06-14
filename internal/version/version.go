package version

const (
	Version = "26.06.13.01"
	Release = false
)

func String() string {
	if Release {
		return Version + "-Release"
	}
	return Version
}
