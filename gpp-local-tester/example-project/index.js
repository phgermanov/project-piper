/**
 * Example application for GPP testing
 */

function main() {
  console.log('Hello from GPP test example!');
  console.log('This is a simple npm project for testing Piper pipeline locally.');
}

if (require.main === module) {
  main();
}

module.exports = { main };
