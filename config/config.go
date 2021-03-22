package config

type Conf struct {
	Region          string `default:"eu-central-1"`
	Bucket          string `required:"true"`
	ChunkFileSizeMB string `split_words:"true" default:"1"`
	MaxFileCount    int    `split_words:"true" default:"10"`
	Host            string `default:"localhost"`
	Port            string `default:"3306"`
	Username        string `default:"root"`
	Password        string `default:""`
	OutputDir       string `default:"output"`
}
