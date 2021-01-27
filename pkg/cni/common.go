package cni

// RegisterCnis - Returns a map with the supported CNI plugins
func RegisterCnis() map[string]Cnier {
	return map[string]Cnier{
		"ovs":         NewOvsCni(),
		"host-device": NewHostDevCni(),
	}
}

// GetKernelSupportedCnis -  Returns a list with the supported Kernel CNIs
func GetKernelSupportedCnis() []string {
	return []string{"ovs", "host-device"}
}
