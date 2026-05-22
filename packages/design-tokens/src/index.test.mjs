import assert from "node:assert/strict";
import test from "node:test";
import { primaryColor, tokens } from "./index.mjs";

test("design tokens preserve legacy Infinitech blue", () => {
  assert.equal(primaryColor, "#009bf5");
  assert.equal(tokens.brand.logoSvg, "assets/brand/logo.svg");
});

