package bot

// serverList contains the list of all servers which are currently playing music
var serverList = ServersState{servers: make(map[string]*ServerState)}
