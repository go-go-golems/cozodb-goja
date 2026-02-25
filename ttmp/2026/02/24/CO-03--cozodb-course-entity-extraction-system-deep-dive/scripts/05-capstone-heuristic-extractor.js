/**
 * Capstone Exercise: Heuristic Tech News Extractor
 *
 * This plugin extracts entities from tech news using simple regex/heuristics.
 * No LLM required -- demonstrates the plugin contract pattern.
 *
 * Usage:
 *   cozo-relationship-js-runner extract scripts/05-capstone-heuristic-extractor.js \
 *       path/to/article.txt
 */

module.exports = {
    apiVersion: "cozo.extractor/v1",
    kind: "extractor",
    id: "capstone.heuristic-extractor",
    name: "Capstone Heuristic Tech News Extractor",

    create: function(ctx) {
        console.log("Heuristic extractor created for run:", ctx.runId);

        return {
            run: function(input) {
                var text = input.transcript || "";

                // --- Extract company-like names ---
                var companyPattern = /([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)/g;
                var companies = {};
                var match;
                while ((match = companyPattern.exec(text)) !== null) {
                    var name = match[1];
                    // Filter: at least 2 chars, not common words
                    var commonWords = ['The', 'In', 'At', 'By', 'For', 'And', 'But', 'Or', 'As', 'Is', 'It', 'To'];
                    if (name.length >= 3 && commonWords.indexOf(name) === -1) {
                        var id = name.toLowerCase().replace(/\s+/g, '_');
                        if (!companies[id]) {
                            companies[id] = { id: id, name: name, mentions: 0 };
                        }
                        companies[id].mentions++;
                    }
                }

                // --- Extract dollar amounts ---
                var moneyPattern = /\$([0-9,.]+)\s*(billion|million|trillion)/gi;
                var amounts = [];
                while ((match = moneyPattern.exec(text)) !== null) {
                    var value = parseFloat(match[1].replace(/,/g, ''));
                    var unit = match[2].toLowerCase();
                    var multiplier = unit === 'trillion' ? 1e12 :
                                     unit === 'billion' ? 1e9 : 1e6;
                    amounts.push({
                        raw: match[0],
                        value_usd: value * multiplier,
                        context: text.substring(
                            Math.max(0, match.index - 60),
                            Math.min(text.length, match.index + match[0].length + 60)
                        ).trim()
                    });
                }

                // --- Extract dates ---
                var months = ['January', 'February', 'March', 'April', 'May', 'June',
                              'July', 'August', 'September', 'October', 'November', 'December'];
                var datePattern = new RegExp('(' + months.join('|') + ')\\s+(\\d{4})', 'g');
                var dates = [];
                while ((match = datePattern.exec(text)) !== null) {
                    var monthIdx = months.indexOf(match[1]) + 1;
                    var monthStr = monthIdx < 10 ? '0' + monthIdx : '' + monthIdx;
                    dates.push({
                        raw: match[0],
                        normalized: match[2] + '-' + monthStr
                    });
                }

                // --- Extract person names (Title + Name pattern) ---
                var personPattern = /(?:(?:Mr|Ms|Mrs|Dr|Prof|CEO|CTO|COO|CFO)\.?\s+)?([A-Z][a-z]+\s+[A-Z][a-z]+)/g;
                var persons = {};
                while ((match = personPattern.exec(text)) !== null) {
                    var personName = match[1];
                    var personId = personName.toLowerCase().replace(/\s+/g, '_');
                    if (!persons[personId]) {
                        persons[personId] = {
                            id: personId,
                            name: personName,
                            mentions: 0,
                            contexts: []
                        };
                    }
                    persons[personId].mentions++;
                    var ctx_start = Math.max(0, match.index - 40);
                    var ctx_end = Math.min(text.length, match.index + match[0].length + 40);
                    if (persons[personId].contexts.length < 3) {
                        persons[personId].contexts.push(text.substring(ctx_start, ctx_end).trim());
                    }
                }

                // --- Build result ---
                var companyList = [];
                for (var cid in companies) {
                    if (companies[cid].mentions >= 2) {
                        companyList.push(companies[cid]);
                    }
                }
                companyList.sort(function(a, b) { return b.mentions - a.mentions; });

                var personList = [];
                for (var pid in persons) {
                    personList.push(persons[pid]);
                }
                personList.sort(function(a, b) { return b.mentions - a.mentions; });

                return JSON.stringify({
                    companies: companyList,
                    persons: personList,
                    financial_amounts: amounts,
                    dates_mentioned: dates,
                    stats: {
                        total_words: text.split(/\s+/).length,
                        total_sentences: text.split(/[.!?]+/).filter(function(s) { return s.trim().length > 0; }).length,
                        companies_found: companyList.length,
                        persons_found: personList.length,
                        amounts_found: amounts.length,
                        dates_found: dates.length
                    }
                }, null, 2);
            }
        };
    }
};
