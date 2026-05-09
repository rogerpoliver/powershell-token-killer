module.exports = {
  extends: ["@commitlint/config-conventional"],
  rules: {
    "type-enum": [
      2,
      "always",
      [
        "feat",     // new feature
        "fix",      // bug fix
        "docs",     // documentation only
        "style",    // formatting, no logic change
        "refactor", // code change, no feat/fix
        "perf",     // performance improvement
        "test",     // adding/fixing tests
        "build",    // build system or dependencies
        "ci",       // CI configuration
        "chore",    // other changes that don't modify src/test
        "revert",   // revert a previous commit
      ],
    ],
    "subject-case": [2, "always", "lower-case"],
    "header-max-length": [2, "always", 100],
  },
};
