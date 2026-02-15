// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract Moltgit is ERC721, Ownable {
    uint256 private _nextTokenId;
    string private _metadataURI;

    constructor(
        string memory metadataURI_
    ) ERC721("Moltgit", "MGIT") Ownable(msg.sender) {
        _metadataURI = metadataURI_;
        // lets mint one token to the deployer
        _safeMint(msg.sender, _nextTokenId);
        _nextTokenId++;
    }

    // =========================
    // ADMIN FUNCTIONS
    // =========================

    function mint(address agent) external onlyOwner returns (uint256) {
        require(agent != address(0), "Invalid agent address");

        uint256 tokenId = _nextTokenId;
        _nextTokenId++;

        _safeMint(agent, tokenId);

        return tokenId;
    }

    function setMetadataURI(string memory newURI) external onlyOwner {
        _metadataURI = newURI;
    }

    // =========================
    // METADATA
    // =========================

    function tokenURI(
        uint256 tokenId
    ) public view override returns (string memory) {
        require(_ownerOf(tokenId) != address(0), "Token does not exist");
        return _metadataURI;
    }

    // =========================
    // VIEW
    // =========================

    function totalMinted() external view returns (uint256) {
        return _nextTokenId;
    }
}
