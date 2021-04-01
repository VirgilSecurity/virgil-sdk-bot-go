package webhook

import (
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"

	"github.com/VirgilSecurity/virgil-commkit-go/crypto"
	"github.com/VirgilSecurity/virgil-commkit-go/crypto/wrapper/foundation"
	"github.com/VirgilSecurity/virgil-commkit-go/crypto/wrapper/sdk/core"

	"github.com/VirgilSecurity/virgil-sdk-bot-go/storage"
)

var cryptoImpl = &crypto.Crypto{}

var random foundation.Random

func init() {
	rnd := foundation.NewCtrDrbg()

	if err := rnd.SetupDefaults(); err != nil {
		panic(fmt.Errorf("virgil crypto cannot initialize random generator: %w", err))
	}
	random = rnd
}

type Client struct {
	PrivateKey foundation.PrivateKey
	URL        string
	Host       string
	Identity   string
	Token      string
	Storage    storage.Storage
}

func NewClient(url string, storage storage.Storage) (*Client, error) {

	domain, identity, token, err := ParseURL(url)
	if err != nil {
		return nil, err
	}

	return &Client{
		URL:      url,
		Host:     domain,
		Identity: identity,
		Token:    token,
		Storage:  storage,
	}, nil
}

func (c *Client) Init() error {

	if c.Storage == nil {
		return errors.New("storage not initialized")
	}

	if c.Storage.Exists("card") && c.Storage.Exists("key") {
		return nil //TODO LOAD
	}

	sk, err := cryptoImpl.GenerateKeypair()
	if err != nil {
		return err
	}

	mgr := sdk_core.NewCardManager()

	mgr.SetRandom(random)

	rawCard, err := mgr.GenerateRawCard(c.Identity, sk.Unwrap())
	if err != nil {
		return err
	}
	jsonCard := rawCard.ExportAsJson()
	strCard := jsonCard.AsStr()

	req := sdk_core.NewHttpRequestWithBody("POST", c.Host+"/api/v1/webhook/publish-card", []byte(strCard))
	auth := c.Identity + ":" + c.Token
	req.SetAuthHeaderValueFromTypeAndCredentials("Basic", base64.StdEncoding.EncodeToString([]byte(auth)))

	resp, err := sdk_core.VirgilHttpClientSend(req)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("http error %d %s", resp.StatusCode(), resp.Body())
	}

	card := resp.Body()

	if err := c.Storage.Store("card", card); err != nil {
		return err
	}

	skBytes, err := cryptoImpl.ExportPrivateKey(sk)
	if err != nil {
		return err
	}

	if err := c.Storage.Store("key", skBytes); err != nil {
		return err
	}
	c.PrivateKey = sk.Unwrap()
	return nil
}

func (c *Client) SendMessage(text string) error {
	return errors.New("not implemented")
}

func ParseURL(url string) (domain, identity, token string, err error) {
	reg := regexp.MustCompile("(https:/\\/.*?)\\/([a-f0-9]{16})\\/([a-f0-9]{32})")
	if err != nil {
		return
	}
	res := reg.FindStringSubmatch(url)
	if len(res) != 4 {
		err = errors.New("invalid url")
		return
	}
	return res[1], res[2], res[3], nil
}
