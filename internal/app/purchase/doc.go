// Package purchase holds the customer write use cases: buying a single available
// pet and checking out several pets atomically. Each wraps the race-safe
// repository operation and invalidates the cached available-pets listing for
// every affected store, so a sold pet drops off the catalog on the next read.
// The customer identity is supplied by the caller from the authenticated
// principal, never from client input.
package purchase
