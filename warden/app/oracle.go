package app

import (
	"context"
	"math/big"
	"time"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	slinkytypes "github.com/skip-mev/slinky/pkg/types"

	oraclepreblock "github.com/skip-mev/slinky/abci/preblock/oracle"
	"github.com/skip-mev/slinky/abci/proposals"
	"github.com/skip-mev/slinky/abci/strategies/aggregator"
	compression "github.com/skip-mev/slinky/abci/strategies/codec"
	"github.com/skip-mev/slinky/abci/strategies/currencypair"
	"github.com/skip-mev/slinky/abci/ve"
	slinkyaggregator "github.com/skip-mev/slinky/aggregator"
	oracleconfig "github.com/skip-mev/slinky/oracle/config"
	"github.com/skip-mev/slinky/pkg/math/voteweighted"
	oracleclient "github.com/skip-mev/slinky/service/clients/oracle"
	servicemetrics "github.com/skip-mev/slinky/service/metrics"
	oraclekeeper "github.com/skip-mev/slinky/x/oracle/keeper"
)

func (app *App) initializeOracles(appOpts types.AppOptions) {
	app.setupMempool()
	baseProposalHandler := app.baseProposalHandler()
	app.setupSlinkyClient(appOpts)
	app.setupABCILifecycle(
		baseProposalHandler.PrepareProposalHandler(),
		baseProposalHandler.ProcessProposalHandler(),
	)
}

func (app *App) setupMempool() {
	// force NoOpMempool as required by x/evm
	noopMempool := mempool.NoOpMempool{}
	app.SetMempool(noopMempool)
}

func (app *App) setupSlinkyClient(appOpts types.AppOptions) {
	chainID := app.ChainID()
	c, err := NewSlinkyClient(app.Logger(), chainID, appOpts, app.StakingKeeper, app.OracleKeeper)
	if err != nil {
		panic(err)
	}
	app.slinkyClient = c
}

func (app *App) baseProposalHandler() *baseapp.DefaultProposalHandler {
	mempool := app.Mempool()
	return baseapp.NewDefaultProposalHandler(mempool, app)
}

type SlinkyClient struct {
	logger         log.Logger
	cfg            oracleconfig.AppConfig
	metrics        servicemetrics.Metrics
	client         oracleclient.OracleClient
	veCodec        compression.VoteExtensionCodec
	extCommitCodec compression.ExtendedCommitCodec
	stakingKeeper  *stakingkeeper.Keeper
	oracleKeeper   *oraclekeeper.Keeper
}

func NewSlinkyClient(
	logger log.Logger,
	chainID string,
	appOpts types.AppOptions,
	stakingKeeper *stakingkeeper.Keeper,
	oracleKeeper *oraclekeeper.Keeper,
) (*SlinkyClient, error) {
	// Read general config from app-opts, and construct oracle service.
	cfg, err := oracleconfig.ReadConfigFromAppOpts(appOpts)
	if err != nil {
		return nil, err
	}

	// If app level instrumentation is enabled, then wrap the oracle service with a metrics client
	// to get metrics on the oracle service (for ABCI++). This will allow the instrumentation to track
	// latency in VerifyVoteExtension requests and more.
	metrics, err := servicemetrics.NewMetricsFromConfig(cfg, chainID)
	if err != nil {
		return nil, err
	}

	// Create the oracle service.
	client, err := oracleclient.NewClientFromConfig(
		cfg,
		logger.With("client", "oracle"),
		metrics,
	)
	if err != nil {
		return nil, err
	}

	// Connect to the oracle service (default timeout of 5 seconds).
	go func() {
		if err := client.Start(context.Background()); err != nil {
			logger.Error("failed to start oracle client", "err", err)
			// It's okay to panic here because an error here means that the
			// address was misconfigured.
			// In case the oracle server is not running, this won't be an error
			// and the application will continue to run as expected.
			panic(err)
		}

		logger.Info("started oracle client", "address", cfg.OracleAddress)
	}()

	return &SlinkyClient{
		logger:  logger,
		cfg:     cfg,
		metrics: metrics,
		client:  client,
		veCodec: compression.NewCompressionVoteExtensionCodec(
			compression.NewDefaultVoteExtensionCodec(),
			compression.NewZLibCompressor(),
		),
		extCommitCodec: compression.NewCompressionExtendedCommitCodec(
			compression.NewDefaultExtendedCommitCodec(),
			compression.NewZStdCompressor(),
		),
		stakingKeeper: stakingKeeper,
		oracleKeeper:  oracleKeeper,
	}, nil
}

func (sc *SlinkyClient) ProposalHandler(
	prepareProposalHandler sdk.PrepareProposalHandler,
	processProposalHandler sdk.ProcessProposalHandler,
) *proposals.ProposalHandler {
	return proposals.NewProposalHandler(
		sc.logger,
		prepareProposalHandler,
		processProposalHandler,
		ve.NewDefaultValidateVoteExtensionsFn(sc.stakingKeeper),
		sc.veCodec,
		compression.NewCompressionExtendedCommitCodec(
			compression.NewDefaultExtendedCommitCodec(),
			compression.NewZStdCompressor(),
		),
		currencypair.NewDeltaCurrencyPairStrategy(sc.oracleKeeper),
		sc.metrics,
	)
}

func (sc *SlinkyClient) AggregatorFn() slinkyaggregator.AggregateFnFromContext[string, map[slinkytypes.CurrencyPair]*big.Int] {
	// Create the aggregation function that will be used to aggregate oracle data
	// from each validator.
	return voteweighted.MedianFromContext(
		sc.logger,
		sc.stakingKeeper,
		voteweighted.DefaultPowerThreshold,
	)
}

func (sc *SlinkyClient) PreblockHandler() *oraclepreblock.PreBlockHandler {
	return oraclepreblock.NewOraclePreBlockHandler(
		sc.logger,
		sc.AggregatorFn(),
		sc.oracleKeeper,
		sc.metrics,
		currencypair.NewDeltaCurrencyPairStrategy(sc.oracleKeeper),
		sc.veCodec,
		sc.extCommitCodec,
	)
}

func (sc *SlinkyClient) VoteExtensionHandler() *ve.VoteExtensionHandler {
	cps := currencypair.NewDeltaCurrencyPairStrategy(sc.oracleKeeper)
	return ve.NewVoteExtensionHandler(
		sc.logger,
		sc.client,
		time.Second,
		cps,
		sc.veCodec,
		aggregator.NewOraclePriceApplier(
			aggregator.NewDefaultVoteAggregator(
				sc.logger,
				sc.AggregatorFn(),
				// we need a separate price strategy here, so that we can optimistically apply the latest prices
				// and extend our vote based on these prices
				currencypair.NewDeltaCurrencyPairStrategy(sc.oracleKeeper),
			),
			sc.oracleKeeper,
			sc.veCodec,
			sc.extCommitCodec,
			sc.logger,
		),
		sc.metrics,
	)
}

func (app *App) setupABCILifecycle(
	basePrepareProposalHandler sdk.PrepareProposalHandler,
	baseProcessProposalHandler sdk.ProcessProposalHandler,
) {
	slinkyProposalHandler := app.slinkyClient.ProposalHandler(
		basePrepareProposalHandler,
		baseProcessProposalHandler,
	)
	slinkyPreBlockHandler := app.slinkyClient.PreblockHandler()
	slinkyVEHandler := app.slinkyClient.VoteExtensionHandler()

	app.SetPrepareProposal(slinkyProposalHandler.PrepareProposalHandler())
	app.SetProcessProposal(slinkyProposalHandler.ProcessProposalHandler())
	app.SetPreBlocker(slinkyPreBlockHandler.WrappedPreBlocker(app.ModuleManager))
	app.SetExtendVoteHandler(slinkyVEHandler.ExtendVoteHandler())
	app.SetVerifyVoteExtensionHandler(slinkyVEHandler.VerifyVoteExtensionHandler())
}

type AppUpgrade struct {
	Name         string
	Handler      upgradetypes.UpgradeHandler
	StoreUpgrade storetypes.StoreUpgrades
}
