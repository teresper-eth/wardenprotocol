import FaucetButton from "./faucet-button";
import { Button } from "@/components/ui/button";
import { AlertCircle, ChevronsUpDown, Copy } from "lucide-react";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "@/components/ui/popover";
import AddressAvatar from "./address-avatar";
import { useAddressContext } from "@/def-hooks/useAddressContext";
// import useKeplr from "@/def-hooks/useKeplr";
import { useAsset } from "@/def-hooks/useAsset";
import { useDispatchWalletContext } from "../def-hooks/walletContext";

export function ConnectWallet() {
	// const { connectToKeplr, signOut } = useKeplr();
	const { signOut, getActiveWallet } = useDispatchWalletContext();
	const { address } = useAddressContext();

	const { balance } = useAsset("uward");
	const ward = parseInt(balance?.amount || "0") / 10 ** 6;

	const activeWallet = getActiveWallet();

	return (
		<Popover>
			<PopoverTrigger asChild>
				{address ? (
					<Button
						asChild
						variant="outline"
						role="combobox"
						className="justify-between cursor-pointer border-border h-16 border-t-0 border-b-0 rounded-none gap-4 min-w-0 hover:bg-muted hover:border-b-accent hover:border-b-2"
					>
						<div>
							<AddressAvatar seed={address} disableTooltip />
							<div className="flex flex-col text-left text-xs">
								<span className="block text-sm truncate">
									{address.slice(0, 8) +
										"..." +
										address.slice(-8)}
								</span>
								<span className="block text-sm truncate">
									{ward.toFixed(2)} WARD
								</span>
							</div>
							<ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
						</div>
					</Button>
				) : (
					<Button
						variant="outline"
						role="combobox"
						className="justify-between cursor-pointer h-16 border-t-0 border-b-0 rounded-none gap-4 min-w-0 hover:bg-muted hover:border-b-accent hover:border-b-2"
					>
						<div>
							<AlertCircle className="ml-2 h-8 w-8 shrink-0" />
						</div>
						<div className="flex flex-col text-left text-xs">
							<span>Not Connected</span>
							<span>Connect Wallet</span>
						</div>
						<ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
					</Button>
				)}
			</PopoverTrigger>

			{address && (
				<PopoverContent className="w-80 rounded-t-none border-t-0 -translate-y-1">
					<div className="grid gap-4">
						<div className="flex flex-row text-left text-xs gap-2 justify-between items-center">
							<span className="block text-base">
								{address.slice(0, 12) +
									"..." +
									address.slice(-12)}
							</span>
							<span>
								<Copy
									className="h-4 w-4 cursor-pointer"
									onClick={() =>
										navigator.clipboard.writeText(address)
									}
								/>
							</span>
						</div>
						<div className="border rounded-lg">
							<div className="px-6 py-3 text-sm border-b flex justify-between">
								<span>Wallet</span>
								<span>{activeWallet?.name || ""}</span>
							</div>
							<div className="px-6 py-3 text-sm flex justify-between">
								<span>Balance</span>
								<span>{ward.toFixed(2)} WARD</span>
							</div>
						</div>
						<div className="flex flex-col gap-4 flex-grow">
							<FaucetButton />
							<Button
								onClick={() => signOut()}
								size="sm"
								className="mx-auto w-full"
								variant={"destructive"}
							>
								Disconnect Wallet
							</Button>
						</div>
					</div>
				</PopoverContent>
			)}
		</Popover>
	);
}