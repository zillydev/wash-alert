# Wash Alert

## A real-time washing machine status tracker.

- Made for learning Go and Vue.js.
- Faced problem at hostel of repeatedly needing to check for empty washing machines, and also not knowing about malfunctioning machines.
- Made a basic prototype for a washing machine status tracker in 3 days.
- User flow:
    - Users can check the status of washing machines, and book an empty machine with a specified timer.
    - The app will (or was supposed to) send a notification to the user if the timer runs out.
    - The user can tap the “Collect” button after collecting the clothes, which will mark the machine empty again.
- Used WebSocket for real-time communication.
- Tried to use Web Push notifications and service worker for real-time notifications, realised it requires registering app for access on Windows Push Notification Service.
- Made the app installable using PWA for notifications, but later realised PWA notifications don’t work on iOS devices.
- 
