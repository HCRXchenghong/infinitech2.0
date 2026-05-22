import { createRealtimeServer } from "./server.mjs";

const port = Number(process.env.PORT || 9898);
createRealtimeServer().listen(port, () => {
  console.log(`realtime-gateway listening on ${port}`);
});

