#!/bin/bash
# Mock Piper binary for testing the GPP local testing tool
# This simulates Piper execution without requiring the actual Go binary

STEP_NAME=""
FLAGS=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        version)
            echo "Piper version local-dev (mock)"
            echo "Git commit: mock-commit-hash"
            exit 0
            ;;
        artifactPrepareVersion|npmExecuteScripts|checkIfStepActive)
            STEP_NAME="$1"
            shift
            FLAGS="$@"
            break
            ;;
        *)
            shift
            ;;
    esac
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Mock Piper: $STEP_NAME"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Flags: $FLAGS"
echo ""

# Simulate step execution
case $STEP_NAME in
    artifactPrepareVersion)
        echo "Preparing artifact version..."
        mkdir -p .pipeline/commonPipelineEnvironment
        echo "version=1.0.0-mock" > .pipeline/commonPipelineEnvironment/artifactVersion
        echo "✓ Version prepared: 1.0.0-mock"
        ;;

    npmExecuteScripts)
        echo "Executing npm scripts..."
        if [[ "$FLAGS" == *"--install"* ]]; then
            echo "  ➜ npm install"
            npm install || echo "  ⚠ npm install failed (continuing)"
        fi
        if [[ "$FLAGS" == *"ci-build"* ]]; then
            echo "  ➜ npm run ci-build"
            npm run ci-build || echo "  ⚠ npm run ci-build failed (continuing)"
        fi
        echo "✓ npm scripts executed"
        ;;

    checkIfStepActive)
        echo "Checking active steps..."
        mkdir -p .pipeline
        echo '{"Build": true}' > .pipeline/stage_out.json
        echo '{"Build": {"artifactPrepareVersion": true, "npmExecuteScripts": true}}' > .pipeline/step_out.json
        echo "✓ Active steps determined"
        ;;

    *)
        echo "Executing step: $STEP_NAME"
        echo "✓ Step completed successfully"
        ;;
esac

echo ""
exit 0
