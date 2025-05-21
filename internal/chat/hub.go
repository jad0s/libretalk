package chat

import "github.com/jad0s/libretalk/internal/types"

// connections maps username -> all active connections for that user.
var connections = make(map[string][]types.ConnectionInfo)
