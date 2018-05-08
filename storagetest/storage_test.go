package storagetest

import "github.com/glynternet/go-accounting-storage"

// This line ensures that a Storage pointer can be assigned to a storage.Storage
// variable and, therefore, ensures that *Storage satisfies the storage.Storage
// interface
var _ storage.Storage = &Storage{}
