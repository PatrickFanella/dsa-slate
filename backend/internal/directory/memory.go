package directory

import "context"

type MemoryStore struct {
	Data DirectoryData
}

func NewMemoryStore(data DirectoryData) *MemoryStore {
	return &MemoryStore{Data: NormalizeDirectoryData(data)}
}

func (s *MemoryStore) Directory(ctx context.Context) (DirectoryData, error) {
	return NormalizeDirectoryData(s.Data), nil
}
