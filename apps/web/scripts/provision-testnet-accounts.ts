import {
  BASE_FEE,
  Horizon,
  Keypair,
  Networks,
  Operation,
  TransactionBuilder,
} from "@stellar/stellar-sdk";
import { calculateRiskScore } from "../lib/riskScore";
import { fetchStellarAccount } from "../lib/stellar";

const HORIZON_TESTNET = "https://horizon-testnet.stellar.org";
const FRIENDBOT = "https://friendbot.stellar.org";
const server = new Horizon.Server(HORIZON_TESTNET);

type SetOptions = Parameters<typeof Operation.setOptions>[0];

async function fundWithFriendbot(keypair: Keypair): Promise<void> {
  const response = await fetch(`${FRIENDBOT}?addr=${keypair.publicKey()}`);
  if (!response.ok) {
    throw new Error(`Friendbot no pudo fondear ${keypair.publicKey()}: HTTP ${response.status}`);
  }
}

async function submitOptions(
  source: Keypair,
  options: SetOptions[],
  signers: Keypair[],
): Promise<void> {
  const account = await server.loadAccount(source.publicKey());
  let builder = new TransactionBuilder(account, {
    fee: BASE_FEE,
    networkPassphrase: Networks.TESTNET,
  });

  options.forEach((option) => {
    builder = builder.addOperation(Operation.setOptions(option));
  });

  const transaction = builder.setTimeout(30).build();
  signers.forEach((signer) => transaction.sign(signer));
  await server.submitTransaction(transaction);
}

async function provision(): Promise<void> {
  const healthy = Keypair.random();
  const healthyBackup = Keypair.random();
  const inconsistent = Keypair.random();
  const disabledMaster = Keypair.random();
  const disabledMasterBackup = Keypair.random();

  await Promise.all([healthy, inconsistent, disabledMaster].map(fundWithFriendbot));

  await submitOptions(
    healthy,
    [
      { signer: { ed25519PublicKey: healthyBackup.publicKey(), weight: 1 } },
      { lowThreshold: 1, medThreshold: 2, highThreshold: 2 },
    ],
    [healthy],
  );

  await submitOptions(
    inconsistent,
    [{ lowThreshold: 3, medThreshold: 2, highThreshold: 1 }],
    [inconsistent],
  );

  await submitOptions(
    disabledMaster,
    [
      { signer: { ed25519PublicKey: disabledMasterBackup.publicKey(), weight: 1 } },
      { lowThreshold: 1, medThreshold: 2, highThreshold: 2 },
    ],
    [disabledMaster],
  );
  await submitOptions(
    disabledMaster,
    [{ masterWeight: 0 }],
    [disabledMaster, disabledMasterBackup],
  );

  const examples = await Promise.all(
    [
      ["healthy", healthy.publicKey()],
      ["inconsistent-thresholds", inconsistent.publicKey()],
      ["disabled-master-key", disabledMaster.publicKey()],
    ].map(async ([scenario, accountId]) => {
      const account = await fetchStellarAccount(accountId);
      return {
        scenario,
        accountId,
        thresholds: account.thresholds,
        signers: account.signers.map(({ key, weight, type }) => ({ key, weight, type })),
        risk: calculateRiskScore(account),
      };
    }),
  );

  console.log(JSON.stringify({ network: "testnet", generatedAt: new Date().toISOString(), examples }, null, 2));
  console.error("Claves secretas omitidas. Testnet puede reiniciarse; ejecuta el script para generar ejemplos nuevos.");
}

provision().catch((error: unknown) => {
  console.error(error instanceof Error ? error.message : error);
  process.exitCode = 1;
});
