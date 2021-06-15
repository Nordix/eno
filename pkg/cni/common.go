package cni

// RegisterCnis - Returns a map with the supported CNI plugins
func RegisterCnis() map[string]Cnier {
	return map[string]Cnier{
		"ovs":         NewOvsCni(),
		"host-device": NewHostDevCni(),
	}
}

// RegisterIpams - Returns a map with the supported CNI plugins
func RegisterIpams() map[string]Ipam {
	return map[string]Ipam{
		"whereabouts": NewWhereAboutsIpam(),
	}
}
