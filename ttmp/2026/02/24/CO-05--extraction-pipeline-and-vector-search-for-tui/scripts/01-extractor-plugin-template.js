"use strict";

const gp = require("geppetto");
const { defineExtractorPlugin, wrapExtractorRun } = require("geppetto/plugins");

function extractPayloadFromTurn(turn) {
  const blocks = Array.isArray(turn && turn.blocks) ? turn.blocks : [];
  const text = blocks
    .filter((b) => b && b.payload && typeof b.payload.text === "string")
    .map((b) => b.payload.text)
    .join("\n")
    .trim();

  if (!text) {
    throw new Error("assistant did not return text payload");
  }

  const parsed = JSON.parse(text);
  return {
    persons: Array.isArray(parsed.persons) ? parsed.persons : [],
    relationships: Array.isArray(parsed.relationships) ? parsed.relationships : [],
    behaviors: Array.isArray(parsed.behaviors) ? parsed.behaviors : [],
    events: Array.isArray(parsed.events) ? parsed.events : [],
  };
}

module.exports = defineExtractorPlugin({
  id: "cozo.relationship-extractor.template",
  name: "Cozo Relationship Extractor Template",
  create() {
    return {
      run: wrapExtractorRun((input) => {
        const engine = input.engineOptions
          ? gp.engines.fromConfig(input.engineOptions)
          : gp.engines.fromProfile(input.profile || "", { timeoutMs: input.timeoutMs });

        const session = gp
          .createBuilder()
          .withEngine(engine)
          .useGoMiddleware("systemPrompt", {
            prompt:
              input.prompt ||
              "Extract persons, relationships, behaviors, and events as strict JSON.",
          })
          .buildSession();

        const seed = gp.turns.newTurn({
          blocks: [gp.turns.newUserBlock(input.transcript)],
        });

        const out = session.run(seed, {
          timeoutMs: input.timeoutMs,
          tags: {
            app: "cozo-tui",
            ticket: "CO-05",
          },
        });

        return extractPayloadFromTurn(out);
      }),
    };
  },
});
