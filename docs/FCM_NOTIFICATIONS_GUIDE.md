# FCM Notifications Guide

This document describes the flow and content of push notifications dispatched by the EGP Scraper backend using Firebase Cloud Messaging (FCM). Notifications are sent automatically to clients subscribed to specific currency fluctuations when there is a change in the exchange rate.

## 1. Notification Trigger Logic

The notification logic is executed during the rate scraping and persistence phase (located in `api/src/services/fcmAlerts.ts`).

For every scraped currency rate, the system compares the **New Buy Rate** (`newBuy`) with the **Previous Buy Rate** (`oldBuy`) from the last known database snapshot.

- If the rates are identical (`newBuy === oldBuy`), the process exits silently.
- If `newBuy > oldBuy`, the direction is determined as `up` (Increase).
- If `newBuy < oldBuy`, the direction is determined as `down` (Decrease).

## 2. Topic Routing Format

The backend publishes messages directly to FCM Topics so that mobile clients can easily subscribe or unsubscribe based on user preferences.

The dynamic topic string strictly follows this format (in all lowercase):
`alert_{bankId}_{currencyCode}_{currencySuffix}_{increase|decrease}`

**Examples:**
- Subscribe to USD increase alerts at National Bank of Egypt (NBE):
  `alert_nbe_usd_egp_increase`
- Subscribe to EUR decrease alerts at Commercial International Bank (CIB):
  `alert_cib_eur_egp_decrease`
- Subscribe to International Gold (USD) spot price increases:
  `alert_gold_international_international_usd_increase`

## 3. Notification Payload Content

The payload delivered to Firebase consists strictly of a `data` block. The visual `notification` block is **completely absent**, transforming it into a "Data-Only" or "Silent" push notification. The mobile OS will pass this data directly to the application in the background, allowing the client to dynamically render a custom, localized "Curiosity Gap" notification.

### Data-Only Payload Structure

The `data` object contains strictly typed strings to prevent type casting issues across the FCM Node SDK. The Android/iOS client must intercept these key-values and render the visual alert:

```json
{
  "bank_id": "cib",           // The lowercase identifier of the bank
  "currency_code": "usd",     // The 3-letter currency code
  "direction": "INCREASE",    // "INCREASE" or "DECREASE"
  "type": "PRICE_CHANGE"      // Fixed string representing the event type
}
```

## 4. Error Handling

FCM dispatches are wrapped in a resilient `try/catch` block. If the notification fails (e.g., due to a temporary network issue or invalid Firebase credentials), the error is logged, but the primary data ingestion pipeline does not crash, ensuring the database stays up-to-date regardless of notification delivery status.
