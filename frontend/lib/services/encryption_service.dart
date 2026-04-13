import 'dart:convert';
import 'package:cryptography/cryptography.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class EncryptionService {
  static const _storage = FlutterSecureStorage();
  static const _privateKeyKey = 'chat_private_key';
  
  // X25519 for Key Agreement, AES-256-GCM for symmetric encryption
  final _keyAgreementAlgorithm = X25519();
  final _encryptionAlgorithm = AesGcm.with256bits();

  /// Generates a new keypair if it doesn't exist, returns the Base64 public key.
  Future<String> getOrCreatePublicKey() async {
    final storedKeyStr = await _storage.read(key: _privateKeyKey);
    if (storedKeyStr != null) {
      // Keypair exists
      final privateBytes = base64Decode(storedKeyStr);
      final keyPair = await _keyAgreementAlgorithm.newKeyPairFromSeed(privateBytes);
      final publicKey = await keyPair.extractPublicKey();
      return base64Encode(publicKey.bytes);
    }

    // Generate new KeyPair
    final keyPair = await _keyAgreementAlgorithm.newKeyPair();
    final privateKeyBytes = await keyPair.extractPrivateKeyBytes();
    final publicKey = await keyPair.extractPublicKey();

    await _storage.write(key: _privateKeyKey, value: base64Encode(privateKeyBytes));
    return base64Encode(publicKey.bytes);
  }

  /// Derives the shared AES secret using the local private key and peer's public key.
  Future<SecretKey> _deriveSharedSecret(String peerPublicKeyBase64) async {
    final storedKeyStr = await _storage.read(key: _privateKeyKey);
    if (storedKeyStr == null) {
      throw Exception('Local private key not found');
    }

    final privateBytes = base64Decode(storedKeyStr);
    final localKeyPair = await _keyAgreementAlgorithm.newKeyPairFromSeed(privateBytes);

    final peerBytes = base64Decode(peerPublicKeyBase64);
    final peerPublicKey = SimplePublicKey(peerBytes, type: KeyPairType.x25519);

    final sharedSecret = await _keyAgreementAlgorithm.sharedSecretKey(
      keyPair: localKeyPair,
      remotePublicKey: peerPublicKey,
    );

    // Provide a random nonce for KDHF or fixed "TrueConnectChat"
    final kdf = Hkdf(
      hmac: Hmac.sha256(),
      outputLength: 32,
    );

    final aesSecretKey = await kdf.deriveKey(
      secretKey: sharedSecret,
      nonce: utf8.encode('TrueConnectChat'),
    );

    return aesSecretKey;
  }

  /// Encrypts plaintext message. Format: Base64(Nonce + Ciphertext + MAC)
  Future<String> encryptMessage(String plaintext, String peerPublicKeyBase64) async {
    if (peerPublicKeyBase64.isEmpty) {
       throw Exception('Peer public key is empty, cannot encrypt.');
    }
    
    final secretKey = await _deriveSharedSecret(peerPublicKeyBase64);
    
    final secretBox = await _encryptionAlgorithm.encrypt(
      utf8.encode(plaintext),
      secretKey: secretKey,
    );

    // Combine nonce, ciphertext, and mac into one byte array
    final combined = [
      ...secretBox.nonce,
      ...secretBox.cipherText,
      ...secretBox.mac.bytes,
    ];

    return base64Encode(combined);
  }

  /// Decrypts ciphertext. Expects Base64(Nonce + Ciphertext + MAC)
  Future<String> decryptMessage(String encryptedBase64, String peerPublicKeyBase64) async {
    if (peerPublicKeyBase64.isEmpty) {
       return encryptedBase64; // Can't decrypt without key, return raw
    }

    try {
        final secretKey = await _deriveSharedSecret(peerPublicKeyBase64);
        
        final combined = base64Decode(encryptedBase64);
        
        // Aes-Gcm nonce length is 12 bytes
        final nonce = combined.sublist(0, 12);
        // Mac length is 16 bytes
        final macBytes = combined.sublist(combined.length - 16);
        final cipherText = combined.sublist(12, combined.length - 16);

        final secretBox = SecretBox(cipherText, nonce: nonce, mac: Mac(macBytes));

        final cleartextBytes = await _encryptionAlgorithm.decrypt(
          secretBox,
          secretKey: secretKey,
        );

        return utf8.decode(cleartextBytes);
      } catch (e) {
        debugPrint('Decryption failed: $e');
        return '[Encrypted Message - Key Missing/Failed]';
      }
  }
}
