<?php
/**
 * Upload endpoint for index.html on a PHP webspace: runs
 * "camindo-watermark detect" on the uploaded image and returns its JSON.
 * The image is only read from PHP's temporary upload file and never stored.
 *
 * Copyright (C) 2026 the camindo authors (https://github.com/dirkintelligence)
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

// ---- configuration -------------------------------------------------------
// Binary and key file belong OUTSIDE the document root. The web server can
// override the paths without editing this file, e.g. Apache:
//     SetEnv CAMINDO_WM_BINARY   /opt/watermark/bin/camindo-watermark
//     SetEnv CAMINDO_WM_KEY_FILE /opt/watermark/private/watermark.key
$binary  = getenv('CAMINDO_WM_BINARY') ?: __DIR__ . '/../bin/camindo-watermark';
$keyFile = getenv('CAMINDO_WM_KEY_FILE') ?: __DIR__ . '/../private/watermark.key';
$maxMb   = 30;
// --------------------------------------------------------------------------

header('Content-Type: application/json; charset=utf-8');
header('X-Content-Type-Options: nosniff');
header('Cache-Control: no-store');

function fail(int $code, string $msg): void
{
	http_response_code($code);
	echo json_encode(['status' => 'error', 'error' => $msg]);
	exit;
}

if (($_SERVER['REQUEST_METHOD'] ?? '') !== 'POST') {
	fail(405, 'POST required');
}
$f = $_FILES['image'] ?? null;
if (!is_array($f) || !isset($f['error']) || is_array($f['error'])) {
	fail(400, 'no image uploaded');
}
if ($f['error'] === UPLOAD_ERR_INI_SIZE || $f['error'] === UPLOAD_ERR_FORM_SIZE || $f['size'] > $maxMb * 1048576) {
	fail(413, 'file too large');
}
if ($f['error'] !== UPLOAD_ERR_OK || !is_uploaded_file($f['tmp_name'])) {
	fail(400, 'upload failed');
}
if (!is_file($binary) || !is_executable($binary)) {
	fail(500, 'detector not installed');
}

// at most two detections at the same time (memory / CPU of the webspace)
$lock = null;
for ($i = 0; $i < 2 && !$lock; $i++) {
	// (@: a lock file of another user is simply not usable)
	$h = @fopen(sys_get_temp_dir() . "/camindo-watermark-$i.lock", 'c');
	if ($h && flock($h, LOCK_EX | LOCK_NB)) {
		$lock = $h;
	}
}
if (!$lock) {
	fail(503, 'busy, please try again');
}

$cmd = [$binary, 'detect', '-json', '-key-file', $keyFile, '-in', $f['tmp_name']];
$proc = proc_open($cmd, [1 => ['pipe', 'w'], 2 => ['pipe', 'w']], $pipes);
if (!is_resource($proc)) {
	fail(500, 'detector could not be started');
}
$out = stream_get_contents($pipes[1]);
$err = stream_get_contents($pipes[2]);
fclose($pipes[1]);
fclose($pipes[2]);
$exit = proc_close($proc);
flock($lock, LOCK_UN);

$json = json_decode((string)$out, true);
if (!is_array($json)) {
	error_log('camindo-watermark: exit ' . $exit . ': ' . trim($err));
	fail(500, 'detection failed');
}
if (($json['status'] ?? '') === 'error') {
	http_response_code(422);
}
echo json_encode($json);
