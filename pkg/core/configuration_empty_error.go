package core

func (m *ConfigurationEmptyError) Error() string {
	return "configuration is empty"
}
