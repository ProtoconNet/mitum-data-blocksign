package cmds

import (
	"context"
	"crypto/tls"
	"reflect"
	"sync"
	"time"

	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/key"
	"github.com/spikeekips/mitum/base/node"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/base/seal"
	"github.com/spikeekips/mitum/launch/config"
	"github.com/spikeekips/mitum/launch/pm"
	"github.com/spikeekips/mitum/launch/process"
	"github.com/spikeekips/mitum/network"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/logging"

	"github.com/soonkuk/mitum-data/digest"
)

const (
	ProcessNameDigestAPI      = "digest_api"
	ProcessNameStartDigestAPI = "start_digest_api"
	HookNameSetLocalChannel   = "set_local_channel"
)

var (
	ProcessorDigestAPI      pm.Process
	ProcessorStartDigestAPI pm.Process
)

func init() {
	if i, err := pm.NewProcess(ProcessNameDigestAPI, []string{ProcessNameDigestDatabase}, ProcessDigestAPI); err != nil {
		panic(err)
	} else {
		ProcessorDigestAPI = i
	}

	if i, err := pm.NewProcess(
		ProcessNameStartDigestAPI,
		[]string{ProcessNameDigestDatabase, ProcessNameDigestAPI},
		ProcessStartDigestAPI,
	); err != nil {
		panic(err)
	} else {
		ProcessorStartDigestAPI = i
	}
}

func ProcessStartDigestAPI(ctx context.Context) (context.Context, error) {
	var nt *digest.HTTP2Server
	if err := LoadDigestNetworkContextValue(ctx, &nt); err != nil {
		if xerrors.Is(err, util.ContextValueNotFoundError) {
			return ctx, nil
		}

		return ctx, err
	}

	return ctx, nt.Start()
}

func ProcessDigestAPI(ctx context.Context) (context.Context, error) {
	var design DigestDesign
	if err := LoadDigestDesignContextValue(ctx, &design); err != nil {
		if xerrors.Is(err, util.ContextValueNotFoundError) {
			return ctx, nil
		}

		return ctx, err
	}

	var log logging.Logger
	if err := config.LoadLogContextValue(ctx, &log); err != nil {
		return ctx, err
	}

	var networkLog logging.Logger
	if err := config.LoadNetworkLogContextValue(ctx, &networkLog); err != nil {
		return ctx, err
	}

	if design.Network() == nil {
		log.Debug().Msg("digest api disabled; empty network")

		return ctx, nil
	}

	var st *digest.Database
	if err := LoadDigestDatabaseContextValue(ctx, &st); err != nil {
		log.Debug().Err(err).Msg("digest api disabled; empty database")

		return ctx, nil
	} else if st == nil {
		log.Debug().Msg("digest api disabled; empty database")

		return ctx, nil
	}

	log.Info().
		Str("bind", design.Network().Bind().String()).
		Str("publish", design.Network().ConnInfo().String()).
		Msg("trying to start http2 server for digest API")

	var nt *digest.HTTP2Server
	var certs []tls.Certificate
	if design.Network().Bind().Scheme == "https" {
		certs = design.Network().Certs()
	}

	if sv, err := digest.NewHTTP2Server(
		design.Network().Bind().Host,
		design.Network().ConnInfo().URL().Host,
		certs,
	); err != nil {
		return ctx, err
	} else if err := sv.Initialize(); err != nil {
		return ctx, err
	} else {
		_ = sv.SetLogger(networkLog)

		nt = sv
	}

	return context.WithValue(ctx, ContextValueDigestNetwork, nt), nil
}

func newSendHandler(
	priv key.Privatekey,
	networkID base.NetworkID,
	chans []network.Channel,
) func(interface{}) (seal.Seal, error) {
	return func(v interface{}) (seal.Seal, error) {
		var sl seal.Seal
		switch t := v.(type) {
		case operation.Seal, seal.Seal:
			if s, err := signSeal(v.(seal.Seal), priv, networkID); err != nil {
				return nil, err
			} else if err := s.IsValid(networkID); err != nil {
				return nil, err
			} else {
				sl = s
			}
		case operation.Operation:
			if bs, err := operation.NewBaseSeal(
				priv,
				[]operation.Operation{t},
				networkID,
			); err != nil {
				return nil, xerrors.Errorf("failed to create operation.Seal: %w", err)
			} else if err := bs.IsValid(networkID); err != nil {
				return nil, err
			} else {
				sl = bs
			}
		default:
			return nil, xerrors.Errorf("unsupported message type, %T", t)
		}

		var wg sync.WaitGroup
		wg.Add(len(chans))

		errchan := make(chan error, len(chans))
		for i := range chans {
			go func(i int) {
				defer wg.Done()

				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()

				errchan <- chans[i].SendSeal(ctx, sl)
			}(i)
		}

		wg.Wait()
		close(errchan)

		for err := range errchan {
			if err == nil {
				continue
			}

			return sl, err
		}

		return sl, nil
	}
}

func signSeal(sl seal.Seal, priv key.Privatekey, networkID base.NetworkID) (seal.Seal, error) {
	p := reflect.New(reflect.TypeOf(sl))
	p.Elem().Set(reflect.ValueOf(sl))

	signer := p.Interface().(seal.Signer)

	if err := signer.Sign(priv, networkID); err != nil {
		return nil, err
	}

	return p.Elem().Interface().(seal.Seal), nil
}

func HookSetLocalChannel(ctx context.Context) (context.Context, error) {
	var ln config.LocalNode
	if err := config.LoadConfigContextValue(ctx, &ln); err != nil {
		return ctx, err
	}
	conf := ln.Network()

	var local *node.Local
	if err := process.LoadLocalNodeContextValue(ctx, &local); err != nil {
		return nil, err
	}

	var nodepool *network.Nodepool
	if err := process.LoadNodepoolContextValue(ctx, &nodepool); err != nil {
		return nil, err
	}

	ch, err := process.LoadNodeChannel(conf.ConnInfo(), encs, time.Second*30)
	if err != nil {
		return ctx, err
	}
	if err := nodepool.SetChannel(local.Address(), ch); err != nil {
		return ctx, err
	}

	return ctx, nil
}
