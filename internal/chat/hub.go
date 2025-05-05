package chat

import "libretalk/internal/types"

// connections maps username -> all active connections for that user.
var connections = make(map[string][]types.ConnectionInfo)
