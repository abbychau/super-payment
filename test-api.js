#!/usr/bin/env node

/**
 * Test script for Super Payment API
 * Run with: node test-api.js
 */

const http = require('http');
const https = require('https');

// Configuration
const API_BASE_URL = 'http://localhost:8080';
let authToken = '';

// Helper function to make HTTP requests
function makeRequest(options, data = null) {
    return new Promise((resolve, reject) => {
        const protocol = options.protocol === 'https:' ? https : http;
        
        const req = protocol.request(options, (res) => {
            let body = '';
            
            res.on('data', (chunk) => {
                body += chunk;
            });
            
            res.on('end', () => {
                try {
                    const response = {
                        statusCode: res.statusCode,
                        headers: res.headers,
                        body: body ? JSON.parse(body) : null
                    };
                    resolve(response);
                } catch (error) {
                    resolve({
                        statusCode: res.statusCode,
                        headers: res.headers,
                        body: body
                    });
                }
            });
        });
        
        req.on('error', (error) => {
            reject(error);
        });
        
        if (data) {
            req.write(JSON.stringify(data));
        }
        
        req.end();
    });
}

// Test functions
async function testHealthCheck() {
    console.log('\n🔍 Testing Health Check...');
    
    try {
        const response = await makeRequest({
            hostname: 'localhost',
            port: 8080,
            path: '/health',
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        console.log(`Status: ${response.statusCode}`);
        console.log(`Response:`, JSON.stringify(response.body, null, 2));
        
        if (response.statusCode === 200) {
            console.log('✅ Health check passed!');
            return true;
        } else {
            console.log('❌ Health check failed!');
            return false;
        }
    } catch (error) {
        console.log('❌ Health check error:', error.message);
        return false;
    }
}

async function testUserRegistration() {
    console.log('\n🔍 Testing User Registration...');
    
    // Use timestamp to make email unique for each test run
    const timestamp = Date.now();
    const registrationData = {
        company: {
            corporate_name: "Test Company Inc.",
            representative: "John Doe",
            phone_number: "03-1234-5678",
            postal_code: "100-0001",
            address: "Tokyo, Chiyoda-ku, Example 1-1-1"
        },
        user: {
            full_name: "John Doe",
            email: `john+${timestamp}@testcompany.com`,
            password: "securepassword123"
        }
    };
    
    try {
        const response = await makeRequest({
            hostname: 'localhost',
            port: 8080,
            path: '/api/auth/register',
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        }, registrationData);
        
        console.log(`Status: ${response.statusCode}`);
        console.log(`Response:`, JSON.stringify(response.body, null, 2));
        
        if (response.statusCode === 201 && response.body.token) {
            console.log('✅ User registration successful!');
            authToken = response.body.token;
            return true;
        } else {
            console.log('❌ User registration failed!');
            return false;
        }
    } catch (error) {
        console.log('❌ User registration error:', error.message);
        return false;
    }
}

async function testUserLogin() {
    console.log('\n🔍 Testing User Login...');
    
    const loginData = {
        email: "john@testcompany.com", // Use existing user
        password: "securepassword123"
    };
    
    try {
        const response = await makeRequest({
            hostname: 'localhost',
            port: 8080,
            path: '/api/auth/login',
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        }, loginData);
        
        console.log(`Status: ${response.statusCode}`);
        console.log(`Response:`, JSON.stringify(response.body, null, 2));
        
        if (response.statusCode === 200 && response.body.token) {
            console.log('✅ User login successful!');
            authToken = response.body.token;
            return true;
        } else {
            console.log('❌ User login failed!');
            return false;
        }
    } catch (error) {
        console.log('❌ User login error:', error.message);
        return false;
    }
}

async function testCreateBusinessPartner() {
    console.log('\n🔍 Testing Create Business Partner...');
    
    if (!authToken) {
        console.log('❌ No auth token available!');
        return false;
    }
      const partnerData = {
        company_id: 3, // Add the company ID
        corporate_name: "Test Partner Corp.",
        representative: "Jane Smith",
        phone_number: "03-9876-5432",
        postal_code: "101-0001",
        address: "Tokyo, Chiyoda-ku, Partner Address 2-2-2"
    };
    
    try {
        const response = await makeRequest({
            hostname: 'localhost',
            port: 8080,
            path: '/api/business-partners',
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            }
        }, partnerData);
        
        console.log(`Status: ${response.statusCode}`);
        console.log(`Response:`, JSON.stringify(response.body, null, 2));
        
        if (response.statusCode === 201) {
            console.log('✅ Business partner creation successful!');
            return response.body.data;
        } else {
            console.log('❌ Business partner creation failed!');
            return false;
        }
    } catch (error) {
        console.log('❌ Business partner creation error:', error.message);
        return false;
    }
}

async function testCreateInvoice(businessPartnerId) {
    console.log('\n🔍 Testing Create Invoice...');
    
    if (!authToken) {
        console.log('❌ No auth token available!');
        return false;
    }
    
    if (!businessPartnerId) {
        console.log('❌ No business partner ID available!');
        return false;
    }
    
    const invoiceData = {
        business_partner_id: businessPartnerId,
        payment_amount: 100000.00,
        payment_due_date: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString() // 30 days from now
    };
    
    try {
        const response = await makeRequest({
            hostname: 'localhost',
            port: 8080,
            path: '/api/invoices',
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            }
        }, invoiceData);
        
        console.log(`Status: ${response.statusCode}`);
        console.log(`Response:`, JSON.stringify(response.body, null, 2));
        
        if (response.statusCode === 200) {
            console.log('✅ Invoice creation successful!');
            
            // Verify calculations
            const invoice = response.body.data;
            const expectedFee = invoiceData.payment_amount * 0.04; // 4000.00
            const expectedTax = expectedFee * 0.10; // 400.00
            const expectedTotal = invoiceData.payment_amount + expectedFee + expectedTax; // 104400.00
            
            console.log('\n📊 Calculation Verification:');
            console.log(`Payment Amount: ¥${invoice.payment_amount.toLocaleString()}`);
            console.log(`Fee (4%): ¥${invoice.fee.toLocaleString()} (Expected: ¥${expectedFee.toLocaleString()})`);
            console.log(`Consumption Tax (10% on fee): ¥${invoice.consumption_tax.toLocaleString()} (Expected: ¥${expectedTax.toLocaleString()})`);
            console.log(`Invoice Total: ¥${invoice.invoice_amount.toLocaleString()} (Expected: ¥${expectedTotal.toLocaleString()})`);
            
            if (Math.abs(invoice.fee - expectedFee) < 0.01 && 
                Math.abs(invoice.consumption_tax - expectedTax) < 0.01 && 
                Math.abs(invoice.invoice_amount - expectedTotal) < 0.01) {
                console.log('✅ Invoice calculations are correct!');
            } else {
                console.log('❌ Invoice calculations are incorrect!');
            }
            
            return invoice;
        } else {
            console.log('❌ Invoice creation failed!');
            return false;
        }
    } catch (error) {
        console.log('❌ Invoice creation error:', error.message);
        return false;
    }
}

async function testGetInvoices() {
    console.log('\n🔍 Testing Get Invoices...');
    
    if (!authToken) {
        console.log('❌ No auth token available!');
        return false;
    }
    
    try {
        const response = await makeRequest({
            hostname: 'localhost',
            port: 8080,
            path: '/api/invoices',
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${authToken}`
            }
        });
        
        console.log(`Status: ${response.statusCode}`);
        console.log(`Response:`, JSON.stringify(response.body, null, 2));
        
        if (response.statusCode === 200) {
            console.log('✅ Get invoices successful!');
            return true;
        } else {
            console.log('❌ Get invoices failed!');
            return false;
        }
    } catch (error) {
        console.log('❌ Get invoices error:', error.message);
        return false;
    }
}

async function testGetInvoicesWithFilter() {
    console.log('\n🔍 Testing Get Invoices with Date Filter...');
    
    if (!authToken) {
        console.log('❌ No auth token available!');
        return false;
    }
      const startDate = new Date().toISOString(); // Full RFC3339 format
    const endDate = new Date(Date.now() + 60 * 24 * 60 * 60 * 1000).toISOString(); // 60 days from now
    
    try {
        const response = await makeRequest({
            hostname: 'localhost',
            port: 8080,
            path: `/api/invoices?start_date=${startDate}&end_date=${endDate}&status=unprocessed`,
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${authToken}`
            }
        });
        
        console.log(`Status: ${response.statusCode}`);
        console.log(`Response:`, JSON.stringify(response.body, null, 2));
        
        if (response.statusCode === 200) {
            console.log('✅ Get invoices with filter successful!');
            return true;
        } else {
            console.log('❌ Get invoices with filter failed!');
            return false;
        }
    } catch (error) {
        console.log('❌ Get invoices with filter error:', error.message);
        return false;
    }
}

async function testUnauthorizedAccess() {
    console.log('\n🔍 Testing Unauthorized Access...');
    
    try {
        const response = await makeRequest({
            hostname: 'localhost',
            port: 8080,
            path: '/api/invoices',
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        console.log(`Status: ${response.statusCode}`);
        console.log(`Response:`, JSON.stringify(response.body, null, 2));
        
        if (response.statusCode === 401) {
            console.log('✅ Unauthorized access properly blocked!');
            return true;
        } else {
            console.log('❌ Unauthorized access was not properly blocked!');
            return false;
        }
    } catch (error) {
        console.log('❌ Unauthorized access test error:', error.message);
        return false;
    }
}

// Main test runner
async function runAllTests() {
    console.log('🚀 Starting Super Payment API Tests...');
    console.log('=' .repeat(50));
    
    const results = {
        passed: 0,
        failed: 0,
        total: 0
    };
    
    const tests = [
        { name: 'Health Check', fn: testHealthCheck },
        { name: 'User Registration', fn: testUserRegistration },
        { name: 'User Login', fn: testUserLogin },
        { name: 'Unauthorized Access', fn: testUnauthorizedAccess },
        { name: 'Create Business Partner', fn: testCreateBusinessPartner },
        { name: 'Get Invoices', fn: testGetInvoices },
        { name: 'Get Invoices with Filter', fn: testGetInvoicesWithFilter }
    ];
    
    let businessPartnerId = null;
    
    for (const test of tests) {
        results.total++;
        
        try {
            let result;
            if (test.name === 'Create Business Partner') {
                result = await test.fn();
                if (result && result.id) {
                    businessPartnerId = result.id;
                }
            } else {
                result = await test.fn();
            }
            
            if (result) {
                results.passed++;
            } else {
                results.failed++;
            }
        } catch (error) {
            console.log(`❌ ${test.name} threw an error:`, error.message);
            results.failed++;
        }
        
        // Add a small delay between tests
        await new Promise(resolve => setTimeout(resolve, 500));
    }
    
    // Test invoice creation separately since it needs the business partner ID
    if (businessPartnerId) {
        console.log('\n🔍 Testing Create Invoice...');
        results.total++;
        
        try {
            const result = await testCreateInvoice(businessPartnerId);
            if (result) {
                results.passed++;
            } else {
                results.failed++;
            }
        } catch (error) {
            console.log('❌ Create Invoice threw an error:', error.message);
            results.failed++;
        }
    }
    
    // Print summary
    console.log('\n' + '=' .repeat(50));
    console.log('📊 Test Results Summary:');
    console.log(`✅ Passed: ${results.passed}`);
    console.log(`❌ Failed: ${results.failed}`);
    console.log(`📈 Total: ${results.total}`);
    console.log(`🎯 Success Rate: ${((results.passed / results.total) * 100).toFixed(1)}%`);
    
    if (results.failed === 0) {
        console.log('\n🎉 All tests passed! The API is working correctly.');
    } else {
        console.log('\n⚠️  Some tests failed. Please check the API implementation.');
    }
}

// Run tests if this script is executed directly
if (require.main === module) {
    runAllTests().catch(console.error);
}

module.exports = {
    runAllTests,
    testHealthCheck,
    testUserRegistration,
    testUserLogin,
    testCreateBusinessPartner,
    testCreateInvoice,
    testGetInvoices,
    testGetInvoicesWithFilter,
    testUnauthorizedAccess
};
