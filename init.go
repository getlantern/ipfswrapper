package ipfswrapper

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/getlantern/go-ipfs/assets"
	"github.com/getlantern/go-ipfs/core"
	"github.com/getlantern/go-ipfs/namesys"
	"github.com/getlantern/go-ipfs/repo/config"
	"github.com/getlantern/go-ipfs/repo/fsrepo"
)

const nBitsForKeypair = 2048

func Init(repoRoot string) error {
	var err error
	conf, err := config.Init(ioutil.Discard, nBitsForKeypair)
	if err != nil {
		return err
	}

	if err := fsrepo.Init(repoRoot, conf); err != nil {
		return err
	}

	if err := addDefaultAssets(repoRoot); err != nil {
		return err
	}

	return initializeIpnsKeyspace(repoRoot)
}

func addDefaultAssets(repoRoot string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, err := fsrepo.Open(repoRoot)
	if err != nil { // NB: repo is owned by the node
		return err
	}

	nd, err := core.NewNode(ctx, &core.BuildCfg{Repo: r})
	if err != nil {
		return err
	}
	defer nd.Close()

	dkey, err := assets.SeedInitDocs(nd)
	if err != nil {
		return fmt.Errorf("init: seeding init docs failed: %s", err)
	}
	log.Debugf("init: seeded init docs %s", dkey)
	return nil
}

func initializeIpnsKeyspace(repoRoot string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, err := fsrepo.Open(repoRoot)
	if err != nil { // NB: repo is owned by the node
		return err
	}

	nd, err := core.NewNode(ctx, &core.BuildCfg{Repo: r})
	if err != nil {
		return err
	}
	defer nd.Close()

	err = nd.SetupOfflineRouting()
	if err != nil {
		return err
	}

	return namesys.InitializeKeyspace(ctx, nd.Namesys, nd.Pinning, nd.PrivateKey)
}
