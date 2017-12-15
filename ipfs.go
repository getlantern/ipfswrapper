package ipfswrapper

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"

	"github.com/getlantern/go-ipfs/core"
	"github.com/getlantern/go-ipfs/core/coreunix"
	"github.com/getlantern/go-ipfs/path"
	"github.com/getlantern/go-ipfs/repo/fsrepo"
	uio "github.com/getlantern/go-ipfs/unixfs/io"
	crypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"

	"github.com/getlantern/golog"
)

var log = golog.LoggerFor("flashlight.ipfs")

type Node struct {
	node   *core.IpfsNode
	pk     crypto.PrivKey
	ctx    context.Context
	cancel context.CancelFunc
}

type IpnsEntry struct {
	Name  string
	Value string
}

func Start(repoDir string, pkfile string) (*Node, error) {
	if !fsrepo.IsInitialized(repoDir) {
		log.Debugf("Creating IPFS repo at %v", repoDir)
		if err := Init(repoDir); err != nil {
			return nil, err
		}
	}

	r, err := fsrepo.Open(repoDir)
	if err != nil {
		return nil, err
	}

	var pk crypto.PrivKey
	if pkfile != "" {
		pk, err = GenKeyIfNotExists(pkfile)
		if err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	cfg := &core.BuildCfg{
		Repo:   r,
		Online: true,
	}

	nd, err := core.NewNode(ctx, cfg)

	if err != nil {
		return nil, err
	}
	return &Node{nd, pk, ctx, cancel}, nil
}

func (node *Node) Stop() {
	node.cancel()
}

func (node *Node) Add(content string) (path string, err error) {
	return coreunix.Add(node.node, strings.NewReader(content))
}

func (node *Node) AddFile(fileName string) (path string, err error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	return coreunix.Add(node.node, file)
}

func (node *Node) GetFile(pt string) (io.Reader, error) {
	p := path.Path(pt)
	dn, err := core.Resolve(node.ctx, node.node.Namesys, node.node.Resolver, p)
	if err != nil {
		return nil, err
	}
	return uio.NewDagReader(node.ctx, dn, node.node.DAG)
}

func (node *Node) Get(pt string) (string, error) {
	reader, err := node.GetFile(pt)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, reader)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (node *Node) Publish(p string) (string, error) {
	ref := path.Path(p)
	k := node.node.PrivateKey
	if node.pk != nil {
		k = node.pk
	}
	err := node.node.Namesys.Publish(node.ctx, k, ref)
	if err != nil {
		return "", err
	}

	pid, err := peer.IDFromPrivateKey(k)
	if err != nil {
		return "", err
	}

	return pid.Pretty(), nil
}

func (node *Node) Resolve(name string) (string, error) {
	p, err := node.node.Namesys.ResolveN(node.ctx, name, 1)
	if err != nil {
		return "", err
	}

	return p.String(), nil
}
