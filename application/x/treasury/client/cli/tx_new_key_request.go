package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
	"gitlab.qredo.com/qrdochain/fusionchain/x/treasury/types"
)

var _ = strconv.Itoa(0)

func CmdNewKeyRequest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new-key-request [workspace-id] [keyring-id] [key-type]",
		Short: "Broadcast message NewKeyRequest",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			workspaceID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			keyringID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			var keyType types.KeyType
			switch strings.ToLower(args[2]) {
			case "ecdsa":
				keyType = types.KeyType_KEY_TYPE_ECDSA_SECP256K1
			case "eddsa":
				keyType = types.KeyType_KEY_TYPE_EDDSA_ED25519
			default:
				return fmt.Errorf("invalid key type: %s. Use one of: ecdsa, eddsa", args[1])
			}

			msg := types.NewMsgNewKeyRequest(
				clientCtx.GetFromAddress().String(),
				workspaceID,
				keyringID,
				keyType,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
