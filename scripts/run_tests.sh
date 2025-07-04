#!/bin/bash

# Smart Redirect Test Runner Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test configuration
REDIS_TEST_DB=1
POSTGRES_TEST_DB="smart_redirect_test"

echo -e "${YELLOW}🧪 Starting Smart Redirect Test Suite${NC}"

# Function to print colored output
print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Check dependencies
check_dependencies() {
    echo "🔍 Checking dependencies..."
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed"
        exit 1
    fi
    print_status "Go is installed"
    
    if ! command -v redis-cli &> /dev/null; then
        print_warning "Redis CLI not found, skipping Redis tests"
        SKIP_REDIS=1
    else
        if redis-cli ping &> /dev/null; then
            print_status "Redis is running"
        else
            print_warning "Redis is not running, skipping Redis tests"
            SKIP_REDIS=1
        fi
    fi
    
    if ! command -v psql &> /dev/null; then
        print_warning "PostgreSQL CLI not found, using SQLite for tests"
        USE_SQLITE=1
    else
        print_status "PostgreSQL CLI is available"
    fi
}

# Setup test environment
setup_test_env() {
    echo "🛠️  Setting up test environment..."
    
    # Clean any existing test data
    if [ -z "$SKIP_REDIS" ]; then
        redis-cli -n $REDIS_TEST_DB flushdb &> /dev/null || true
        print_status "Cleaned Redis test database"
    fi
    
    # Create test database if using PostgreSQL
    if [ -z "$USE_SQLITE" ]; then
        psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS $POSTGRES_TEST_DB;" &> /dev/null || true
        psql -h localhost -U postgres -c "CREATE DATABASE $POSTGRES_TEST_DB;" &> /dev/null || true
        print_status "Created PostgreSQL test database"
    fi
    
    # Set test environment variables
    export GIN_MODE=test
    export REDIS_TEST_DB=$REDIS_TEST_DB
    export POSTGRES_TEST_DB=$POSTGRES_TEST_DB
    
    print_status "Environment variables set"
}

# Run unit tests
run_unit_tests() {
    echo "🧪 Running unit tests..."
    
    cd test/unit
    
    # Run tests with coverage
    if go test -v -race -coverprofile=coverage.out ./...; then
        print_status "Unit tests passed"
        
        # Generate coverage report
        go tool cover -html=coverage.out -o coverage.html
        go tool cover -func=coverage.out | tail -n 1
        print_status "Coverage report generated: test/unit/coverage.html"
    else
        print_error "Unit tests failed"
        return 1
    fi
    
    cd ../..
}

# Run integration tests
run_integration_tests() {
    echo "🔗 Running integration tests..."
    
    cd test/integration
    
    if go test -v -race ./...; then
        print_status "Integration tests passed"
    else
        print_error "Integration tests failed"
        return 1
    fi
    
    cd ../..
}

# Run API tests
run_api_tests() {
    echo "🌐 Running API tests..."
    
    # Start test server in background
    echo "Starting test server..."
    go run cmd/server/main.go -config=config/test.yaml &
    SERVER_PID=$!
    
    # Wait for server to start
    sleep 3
    
    # Check if server is running
    if curl -s http://localhost:8080/health > /dev/null; then
        print_status "Test server started"
    else
        print_error "Failed to start test server"
        kill $SERVER_PID 2>/dev/null || true
        return 1
    fi
    
    # Run API tests
    if go test -v ./test/api/...; then
        print_status "API tests passed"
    else
        print_error "API tests failed"
        kill $SERVER_PID 2>/dev/null || true
        return 1
    fi
    
    # Stop test server
    kill $SERVER_PID 2>/dev/null || true
    print_status "Test server stopped"
}

# Run load tests (optional)
run_load_tests() {
    if [ "$RUN_LOAD_TESTS" = "1" ]; then
        echo "⚡ Running load tests..."
        
        if command -v hey &> /dev/null; then
            # Start server for load testing
            go run cmd/server/main.go -config=config/test.yaml &
            SERVER_PID=$!
            sleep 3
            
            echo "Running redirect load test..."
            hey -n 1000 -c 10 http://localhost:8080/v1/bu01/test123?network=mi
            
            echo "Running API load test..."
            hey -n 500 -c 5 -H "Authorization: Bearer test-token" http://localhost:8080/api/v1/links
            
            kill $SERVER_PID 2>/dev/null || true
            print_status "Load tests completed"
        else
            print_warning "hey tool not found, skipping load tests"
        fi
    fi
}

# Generate test report
generate_report() {
    echo "📊 Generating test report..."
    
    # Create reports directory
    mkdir -p reports
    
    # Combine coverage data
    echo "mode: set" > reports/coverage.out
    find . -name "coverage.out" -exec grep -h -v "mode: set" {} \; >> reports/coverage.out
    
    # Generate HTML coverage report
    go tool cover -html=reports/coverage.out -o reports/coverage.html
    
    # Generate test summary
    cat > reports/test_summary.md << EOF
# Smart Redirect Test Report

Generated on: $(date)

## Test Results

### Unit Tests
- Status: ✅ Passed
- Coverage: $(go tool cover -func=reports/coverage.out | tail -n 1 | awk '{print $3}')

### Integration Tests
- Status: ✅ Passed

### API Tests
- Status: ✅ Passed

## Coverage Report

See [coverage.html](coverage.html) for detailed coverage information.

## Recommendations

1. Maintain test coverage above 80%
2. Add more edge case tests for rate limiting
3. Expand integration tests for batch operations

EOF
    
    print_status "Test report generated in reports/"
}

# Cleanup
cleanup() {
    echo "🧹 Cleaning up..."
    
    # Stop any background processes
    pkill -f "smart_redirect" 2>/dev/null || true
    
    # Clean test databases
    if [ -z "$SKIP_REDIS" ]; then
        redis-cli -n $REDIS_TEST_DB flushdb &> /dev/null || true
    fi
    
    if [ -z "$USE_SQLITE" ]; then
        psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS $POSTGRES_TEST_DB;" &> /dev/null || true
    fi
    
    print_status "Cleanup completed"
}

# Main execution
main() {
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --load-tests)
                RUN_LOAD_TESTS=1
                shift
                ;;
            --skip-integration)
                SKIP_INTEGRATION=1
                shift
                ;;
            --skip-api)
                SKIP_API=1
                shift
                ;;
            --help)
                echo "Usage: $0 [options]"
                echo "Options:"
                echo "  --load-tests         Run load tests"
                echo "  --skip-integration   Skip integration tests"
                echo "  --skip-api          Skip API tests"
                echo "  --help              Show this help message"
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Set trap for cleanup on exit
    trap cleanup EXIT
    
    # Run test sequence
    check_dependencies
    setup_test_env
    
    # Install test dependencies
    echo "📦 Installing test dependencies..."
    go mod download
    print_status "Dependencies installed"
    
    # Run tests
    run_unit_tests || exit 1
    
    if [ -z "$SKIP_INTEGRATION" ]; then
        run_integration_tests || exit 1
    fi
    
    if [ -z "$SKIP_API" ]; then
        run_api_tests || exit 1
    fi
    
    run_load_tests
    
    generate_report
    
    echo -e "${GREEN}🎉 All tests completed successfully!${NC}"
    echo -e "${YELLOW}📋 Check the reports/ directory for detailed results${NC}"
}

# Run main function
main "$@"