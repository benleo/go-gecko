package network

type NetConfig struct {
	Address      string `toml:"networkAddress"`
	ReadTimeout  string `toml:"readTimeout"`
	WriteTimeout string `toml:"writeTimeout"`
	BufferSize   uint   `toml:"bufferSize"`
}
