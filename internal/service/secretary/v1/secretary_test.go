package secretary

import (
	"encoding/hex"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type SecretaryTestSuite struct {
	suite.Suite
	secretary *Secretary
	config    *config.SecretConfig
}

func (suite *SecretaryTestSuite) SetupTest() {
	suite.config, _ = config.NewSecretConfig()
	suite.config.UserKey = "jds__63h3_7ds"
	suite.secretary, _ = NewSecretaryService(suite.config)
}

func TestSecretaryTestSuite(t *testing.T) {
	suite.Run(t, new(SecretaryTestSuite))
}

func (suite *SecretaryTestSuite) TestEncode() {
	tests := []struct {
		name             string
		data             string
		expectedEncoding string
	}{
		{
			name:             "sample 1",
			data:             "sample text string",
			expectedEncoding: "c277fd4361e8c0e81e90bc030a31621ff6ef71503544154b7f0e29aae1f69dec0a00",
		},
		{
			name:             "sample 2",
			data:             "another integer data piece",
			expectedEncoding: "d078ff4765e892bc1286bc461e206256fce9061c0fffc7ae409a76a2c8fd0933da10a997181b1f89e06e",
		},
	}

	// perform each test
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedEncoding, suite.secretary.Encode(tt.data))
		})
	}
}

func (suite *SecretaryTestSuite) TestDecode() {
	var invalidByteError *hex.InvalidByteError
	tests := []struct {
		name             string
		expectedDecoding string
		data             string
		error            error
	}{
		{
			name:             "sample 1",
			expectedDecoding: "sample text string",
			data:             "c277fd4361e8c0e81e90bc030a31621ff6ef71503544154b7f0e29aae1f69dec0a00",
			error:            nil,
		},
		{
			name:             "sample 2",
			expectedDecoding: "another integer data piece",
			data:             "d078ff4765e892bc1286bc461e206256fce9061c0fffc7ae409a76a2c8fd0933da10a997181b1f89e06e",
			error:            nil,
		},
		{
			name:             "sample 3",
			expectedDecoding: "",
			data:             "non-hex-encoded-data",
			error:            invalidByteError,
		},
		{
			name:             "sample 4",
			expectedDecoding: "",
			data:             "d078ff4765e892bc1286bc461e206256fce9061c0fffc7ae409a76a",
			error:            nil,
		},
	}

	// perform each test
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			res, err := suite.secretary.Decode(tt.data)
			if err != nil {
				assert.ErrorAs(t, err, &tt.error)
			}
			assert.Equal(t, tt.expectedDecoding, res)

		})
	}
}
