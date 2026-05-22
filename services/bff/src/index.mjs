import { createBffServer } from "./server.mjs";

const port = Number(process.env.PORT || 25500);
createBffServer().listen(port, () => {
  console.log(`bff listening on ${port}`);
});

