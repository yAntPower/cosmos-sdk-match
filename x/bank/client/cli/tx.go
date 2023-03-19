package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

var FlagSplit = "split"

// NewTxCmd returns a root CLI command handler for all x/bank transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Bank transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewSendTxCmd(),
		NewMultiSendTxCmd(),
		NewSellTxCmd(),
		NewBuyTxCmd(),
		NewSetOrderPriceCmd(),
	)

	return txCmd
}

// NewSendTxCmd returns a CLI command handler for creating a MsgSend transaction.
func NewSendTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send [from_key_or_address] [to_address] [amount]",
		Short: "Send funds from one account to another.",
		Long: `Send funds from one account to another.
Note, the '--from' flag is ignored as it is implied from [from_key_or_address].
When using '--dry-run' a key name cannot be used, only a bech32 address.
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set(flags.FlagFrom, args[0])
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			toAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgSend(clientCtx.GetFromAddress(), toAddr, coins)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewMultiSendTxCmd returns a CLI command handler for creating a MsgMultiSend transaction.
// For a better UX this command is limited to send funds from one account to two or more accounts.
func NewMultiSendTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multi-send [from_key_or_address] [to_address_1, to_address_2, ...] [amount]",
		Short: "Send funds from one account to two or more accounts.",
		Long: `Send funds from one account to two or more accounts.
By default, sends the [amount] to each address of the list.
Using the '--split' flag, the [amount] is split equally between the addresses.
Note, the '--from' flag is ignored as it is implied from [from_key_or_address].
When using '--dry-run' a key name cannot be used, only a bech32 address.
`,
		Args: cobra.MinimumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set(flags.FlagFrom, args[0])
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoinsNormalized(args[len(args)-1])
			if err != nil {
				return err
			}

			if coins.IsZero() {
				return fmt.Errorf("must send positive amount")
			}

			split, err := cmd.Flags().GetBool(FlagSplit)
			if err != nil {
				return err
			}

			totalAddrs := sdk.NewInt(int64(len(args) - 2))
			// coins to be received by the addresses
			sendCoins := coins
			if split {
				sendCoins = coins.QuoInt(totalAddrs)
			}

			var output []types.Output
			for _, arg := range args[1 : len(args)-1] {
				toAddr, err := sdk.AccAddressFromBech32(arg)
				if err != nil {
					return err
				}

				output = append(output, types.NewOutput(toAddr, sendCoins))
			}

			// amount to be send from the from address
			var amount sdk.Coins
			if split {
				// user input: 1000stake to send to 3 addresses
				// actual: 333stake to each address (=> 999stake actually sent)
				amount = sendCoins.MulInt(totalAddrs)
			} else {
				amount = coins.MulInt(totalAddrs)
			}

			msg := types.NewMsgMultiSend([]types.Input{types.NewInput(clientCtx.FromAddress, amount)}, output)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool(FlagSplit, false, "Send the equally split token amount to each address")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewSellTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sell [quantity] [price]",
		Short: "sell the coin.",
		Long:  `quantity:quantity of coins. price:unit price.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			fromAddr := clientCtx.GetFromAddress()
			num, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				return err
			}

			price, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				return err
			}
			msg := types.NewMsgTransaction(fromAddr, float32(num), float32(price), sdk.SELL)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewBuyTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buy [quantity] [price]",
		Short: "buy the coin.",
		Long:  `quantity:quantity of coins. price:unit price.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			num, err := strconv.ParseFloat(args[0], 32)
			if err != nil {
				return err
			}
			price, err := strconv.ParseFloat(args[1], 32)
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			fromAddr := clientCtx.GetFromAddress()
			msg := types.NewMsgTransaction(fromAddr, float32(num), float32(price), sdk.BUY)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewSetOrderPriceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setOrderPrice [price] ",
		Short: "set OrderPrice.",
		Long:  `set OrderPrice`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			fromAddr := clientCtx.GetFromAddress()

			price, err := strconv.ParseFloat(args[0], 32)
			if err != nil {
				return err
			}
			fmt.Println(fromAddr, price)
			msg := types.NewMsgTransaction(fromAddr, float32(1), float32(price), sdk.PRICE)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
