package versionrouter

import "github.com/RMS-Server/RMS-Gate/internal/config"

// Group represents a version group with its allowed servers.
type Group struct {
	Name          string
	DefaultServer string
	Servers       map[string]struct{}
}

// Router maps protocol versions to server groups.
type Router struct {
	protoToGroup map[int]string
	groups       map[string]*Group
}

// New creates a Router from config.
func New(cfg *config.VersionRouterConfig) *Router {
	r := &Router{
		protoToGroup: make(map[int]string),
		groups:       make(map[string]*Group),
	}
	for _, gc := range cfg.Groups {
		g := &Group{
			Name:          gc.Name,
			DefaultServer: gc.DefaultServer,
			Servers:       make(map[string]struct{}, len(gc.Servers)),
		}
		for _, s := range gc.Servers {
			g.Servers[s] = struct{}{}
		}
		r.groups[gc.Name] = g
		for _, proto := range gc.Protocols {
			r.protoToGroup[proto] = gc.Name
		}
	}
	return r
}

// GroupForProtocol returns the group for the given protocol, or nil if unsupported.
func (r *Router) GroupForProtocol(protocol int) *Group {
	name, ok := r.protoToGroup[protocol]
	if !ok {
		return nil
	}
	return r.groups[name]
}

// IsServerInGroup checks if a server belongs to the named group.
func (r *Router) IsServerInGroup(groupName, serverName string) bool {
	g, ok := r.groups[groupName]
	if !ok {
		return false
	}
	_, ok = g.Servers[serverName]
	return ok
}

// ServersInGroup returns all server names in the named group.
func (r *Router) ServersInGroup(groupName string) []string {
	g, ok := r.groups[groupName]
	if !ok {
		return nil
	}
	servers := make([]string, 0, len(g.Servers))
	for s := range g.Servers {
		servers = append(servers, s)
	}
	return servers
}
