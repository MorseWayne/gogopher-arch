const mode = process.argv[2]

if (mode !== 'base' && mode !== 'dev') {
  throw new Error('usage: assert-compose-exposure.mjs <base|dev>')
}

let source = ''
for await (const chunk of process.stdin) source += chunk

const { services = {} } = JSON.parse(source)
const ports = (name) => services[name]?.ports ?? []
const publishedServices = Object.entries(services)
  .filter(([, service]) => (service.ports ?? []).length > 0)
  .map(([name]) => name)
  .sort()

function assert(condition, message) {
  if (!condition) throw new Error(message)
}

function assertServiceSet(expected) {
  const actual = publishedServices.join(',')
  assert(
    actual === [...expected].sort().join(','),
    `${mode}: expected published services [${expected.join(', ')}], got [${publishedServices.join(', ')}]`,
  )
}

function assertLoopback(name) {
  const bindings = ports(name)
  assert(bindings.length > 0, `${mode}: ${name} must publish at least one port`)
  for (const binding of bindings) {
    assert(
      binding.host_ip === '127.0.0.1',
      `${mode}: ${name} target ${binding.target} is not bound to 127.0.0.1`,
    )
  }
}

if (mode === 'base') {
  assertServiceSet(['web'])
  assertLoopback('web')
  for (const name of ['gateway', 'sandbox-engine', 'postgres']) {
    assert(ports(name).length === 0, `base: ${name} must remain internal-only`)
  }
} else {
  assertServiceSet(['gateway', 'postgres', 'sandbox-engine', 'web'])
  for (const name of publishedServices) assertLoopback(name)
}

console.log(`${mode} Compose exposure is loopback-only and matches policy`)
