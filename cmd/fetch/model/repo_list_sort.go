package model

// RepositoryListSort represents the field to sort the repository list by.
type RepositoryListSort int
type RepositoryListSortDirection bool

const (
	RepoListSortCreated RepositoryListSort = iota
	RepoListSortUpdated
	RepoListSortPushed
	RepoListSortFullName
)

// String returns the string representation of a RepositoryListSort.
// The returned string is based on the value of the RepositoryListSort
// and is determined by the map that associates RepositoryListSort values
// with their corresponding string representations.
func (r RepositoryListSort) String() string {
	return map[RepositoryListSort]string{
		RepoListSortCreated:  "created",
		RepoListSortUpdated:  "updated",
		RepoListSortPushed:   "pushed",
		RepoListSortFullName: "full_name",
	}[r]
}
