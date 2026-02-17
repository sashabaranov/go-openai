<?php

use App\Klicktipp\CampaignsProcessFlow;

define('DRUPAL_ROOT', getcwd());
define('DEFAULT_BACKUP_DIR', getcwd() . '/campaignCheckBackup');

// Set missing $_SERVER variables for CLI environment
if (php_sapi_name() === 'cli') {
    $_SERVER['REMOTE_ADDR'] = '127.0.0.1';
    $_SERVER['REQUEST_METHOD'] = 'GET';
    $_SERVER['HTTP_HOST'] = 'localhost';
    $_SERVER['REQUEST_URI'] = '/';
    $_SERVER['SCRIPT_NAME'] = '/index.php';
    $_SERVER['PHP_SELF'] = '/index.php';
    $_SERVER['HTTP_USER_AGENT'] = 'CLI';
}

require_once DRUPAL_ROOT . '/includes/bootstrap.inc';
drupal_bootstrap(DRUPAL_BOOTSTRAP_FULL);

$options = parseCommandLineArgs($argv);
$debugMode = $options['debug'] ?? false;
$chunkSize = $options['chunk-size'] ?? 50;
$backupDir = $options['backup-dir'] ?? DEFAULT_BACKUP_DIR;

// Create backup directory if it doesn't exist
if (!is_dir($backupDir)) {
    if (!mkdir($backupDir, 0755, true)) {
        echo "Error: Could not create backup directory: $backupDir\n";
        exit(1);
    }
}

function parseCommandLineArgs($argv)
{
    $options = [];

    for ($i = 1; $i < count($argv); $i++) {
        $arg = $argv[$i];

        if ($arg === '-d' || $arg === '--debug') {
            $options['debug'] = true;
        } elseif ($arg === '-h' || $arg === '--help') {
            showHelp();
            exit(0);
        } elseif (preg_match('/^--chunk-size=(\d+)$/', $arg, $matches)) {
            $options['chunk-size'] = (int)$matches[1];
        } elseif ($arg === '--chunk-size' && isset($argv[$i + 1])) {
            $options['chunk-size'] = (int)$argv[++$i];
        } elseif (preg_match('/^--backup-dir=(.+)$/', $arg, $matches)) {
            $options['backup-dir'] = $matches[1];
        } elseif ($arg === '--backup-dir' && isset($argv[$i + 1])) {
            $options['backup-dir'] = $argv[++$i];
        }
    }

    return $options;
}

function showHelp()
{
    echo "Campaign Process Flow Analyzer\n";
    echo "Usage: php campaignCheck.php [OPTIONS]\n\n";
    echo "Options:\n";
    echo "  -d, --debug           Enable debug output (shows progress and analysis details)\n";
    echo "  --chunk-size=SIZE     Set chunk size (default: 50)\n";
    echo "  --backup-dir=PATH     Set backup directory (default: ./campaignCheckBackup)\n";
    echo "  -h, --help            Show this help message\n\n";
    echo "Examples:\n";
    echo "  php campaignCheck.php                    # Only broken campaigns\n";
    echo "  php campaignCheck.php -d                 # With debug output\n";
    echo "  php campaignCheck.php --debug --chunk-size=100\n";
    echo "  php campaignCheck.php --backup-dir=/path/to/backup\n";
}

function getCampaignsChunked($chunkSize = 100)
{
    $offset = 0;

    do {
        $campaigns = [];

        $result = db_query(
            "SELECT CampaignID, RelOwnerUSerID, Data FROM {campaigns} c "
            . "WHERE 1 ORDER BY c.CampaignID LIMIT $offset, $chunkSize"
        );

        while ($row = kt_fetch_array($result)) {
            $campaigns[] = $row;
        }

        if (!empty($campaigns)) {
            yield $campaigns;
        }

        $offset += $chunkSize;
    } while (count($campaigns) === $chunkSize);
}

/**
 * Analyzes a process flow to find elements that point to the same next/nextno IDs
 * Excludes elements of type 'goto'
 *
 * @param array $data process flow data
 *
 * @return array Results showing duplicate next/nextno references
 */
function findDuplicateNextReferences($data)
{
    if (!is_array($data)) {
        return ['error' => 'Invalid serialized data'];
    }

    $nextReferences = [];
    $elements = [];

    if (
        isset($data['ProcessFlow']['states']) && is_array(
            $data['ProcessFlow']['states']
        )
    ) {
        foreach ($data['ProcessFlow']['states'] as $element) {
            if (isset($element['id'])) {
                $elements[$element['id']] = $element;
            }
        }
    }

    foreach ($elements as $elementId => $element) {
        if (isset($element['type']) && $element['type'] === 'goto') {
            continue;
        }

        if (
            isset($element['next']) && is_numeric(
                $element['next']
            ) && $element['next'] > 0
        ) {
            $nextId = $element['next'];
            $nextReferences[$nextId][] = [
            'sourceId' => $elementId,
            'sourceType' => $element['type'] ?? 'unknown',
            'referenceType' => 'next',
            'element' => $element
            ];
        }

        if (
            isset($element['nextNo']) && is_numeric(
                $element['nextNo']
            ) && $element['nextNo'] > 0
        ) {
            $nextId = $element['nextNo'];
            $nextReferences[$nextId][] = [
            'sourceId' => $elementId,
            'sourceType' => $element['type'] ?? 'unknown',
            'referenceType' => 'nextNo',
            'element' => $element
            ];
        }
    }

    $duplicates = [];
    foreach ($nextReferences as $nextId => $references) {
        if (count($references) > 1) {
            $duplicates[$nextId] = $references;
        }
    }

    return [
    'duplicates' => $duplicates,
    'elements' => $elements
    ];
}

function createBackupFile($backupDir, $userId, $campaignId, $originalData)
{
    $timestamp = date('Y-m-d_H-i-s');
    $filename = "{$userId}_{$campaignId}_{$timestamp}.txt";
    $filepath = $backupDir . '/' . $filename;

    $content = "UserID: $userId\n";
    $content .= "CampaignID: $campaignId\n";
    $content .= "Timestamp: " . date('Y-m-d H:i:s') . "\n";
    $content .= "Original Data:\n";
    $content .= $originalData;

    return file_put_contents($filepath, $content) !== false;
}

function debugOutput($message, $data = null, $debugMode = false)
{
    if (!$debugMode) {
        return;
    }

    echo "DEBUG: $message\n";

    if ($data !== null) {
        if (is_array($data) || is_object($data)) {
            print_r($data);
        } else {
            echo "$data\n";
        }
    }

    echo "\n";
}

function formatBytes($bytes, $precision = 2)
{
    $units = array('B', 'KB', 'MB', 'GB', 'TB');

    for ($i = 0; $bytes > 1024 && $i < count($units) - 1; $i++) {
        $bytes /= 1024;
    }

    return round($bytes, $precision) . ' ' . $units[$i];
}

$totalProcessed = 0;
$brokenCampaigns = 0;
$chunkCount = 0;

if ($debugMode) {
    echo "Starting campaign analysis...\n";
    echo "Debug mode: ENABLED\n";
    echo "Chunk size: $chunkSize\n";
    echo "Backup directory: $backupDir\n";
    echo str_repeat("=", 80) . "\n\n";
}

foreach (getCampaignsChunked($chunkSize) as $campaignChunk) {
    $chunkCount++;
    $currentChunkSize = count($campaignChunk);

    debugOutput("Processing chunk #$chunkCount with $currentChunkSize campaigns", null, $debugMode);
    debugOutput("Memory usage: " . formatBytes(memory_get_usage()), null, $debugMode);

    foreach ($campaignChunk as $campaign) {
        $totalProcessed++;

        $result = findDuplicateNextReferences(unserialize($campaign['Data']));
        if (!empty($result['duplicates'])) {
            $brokenCampaigns++;

          // Create backup file for broken campaign
            if (createBackupFile($backupDir, $campaign['RelOwnerUSerID'], $campaign['CampaignID'], $campaign['Data'])) {
                debugOutput("Backup created for campaign {$campaign['CampaignID']}", null, $debugMode);
            } else {
                echo "Warning: Failed to create backup for campaign {$campaign['CampaignID']}\n";
            }

            echo "Process Flow Analysis CampaignID: " . $campaign['CampaignID']
            . " UserID: " . $campaign['RelOwnerUSerID'] . "\n";
            echo "============================================================\n\n";

            echo "Found " . count($result['duplicates']) . " elements with duplicate references:\n\n";

            foreach ($result['duplicates'] as $nextId => $references) {
                $targetElement = isset($result['elements'][$nextId])
                ? "'" . ($result['elements'][$nextId]['type'] ?? 'unknown') . "'"
                : "unknown";

                echo "Multiple elements point to ID $nextId ($targetElement):\n";

                foreach ($references as $ref) {
                    $element = $ref['element'];
                    $simplifiedElement = [
                    'id' => $element['id'],
                    'type' => $element['type'] ?? 'unknown',
                    'name' => $element['name']
                    ];

                    if (isset($element['next'])) {
                        $simplifiedElement['next'] = $element['next'];
                    }

                    if (isset($element['nextNo'])) {
                        $simplifiedElement['nextNo'] = $element['nextNo'];
                    }

                    echo "- Element ID: {$ref['sourceId']} (Type: {$ref['sourceType']}) via {$ref['referenceType']}\n";
                    echo "  " . json_encode($simplifiedElement, JSON_PRETTY_PRINT) . "\n";
                }
                echo "\n";
            }

            if ($debugMode) {
                echo str_repeat("-", 80) . "\n";
            }
        } else {
            debugOutput("Campaign {$campaign['CampaignID']}: No duplicates found", null, $debugMode);
        }

        if ($debugMode && $totalProcessed % 100 === 0) {
            echo "DEBUG: Progress - Processed $totalProcessed campaigns, found $brokenCampaigns broken\n\n";
        }
    }

    debugOutput("Completed chunk #$chunkCount", null, $debugMode);
}

echo "\nAnalysis complete!\n";
echo "Total campaigns processed: $totalProcessed\n";
echo "Broken campaigns found: $brokenCampaigns\n";
echo "Total chunks processed: $chunkCount\n";
echo "Backup directory: $backupDir\n";
echo "Final memory usage: " . formatBytes(memory_get_usage()) . "\n";
echo "Peak memory usage: " . formatBytes(memory_get_peak_usage()) . "\n";

if ($brokenCampaigns === 0) {
    echo "\nNo broken campaigns found.\n";
}