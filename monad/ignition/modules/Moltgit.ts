import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export default buildModule("Moltgit", (m) => {
  const metadataURI_ =
    "https://orange-chemical-mole-844.mypinata.cloud/ipfs/bafybeig6bmigqxemis2yne7xc5vbqo6msfhoxhydjq432gtgxy5ca4knw4"; // Dev note: placeholder, we need change to this later

  const moltgit = m.contract("Moltgit", [metadataURI_]);

  return { moltgit };
});
