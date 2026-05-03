//! prompt-optimizer — heuristic prompt compression for LLM token savings.
//!
//! Strips fillers, polite pleasantries, redundant intensifiers, and verbose
//! phrasings without changing the intent. Deterministic, no network, no LLM
//! call. Typical savings on chatty prompts: 10–25%.

use extism_pdk::*;
use serde::{Deserialize, Serialize};

#[derive(Deserialize)]
struct OptimizeInput {
    prompt: String,
    /// Optional: keep_polite=true skips the politeness/pleasantry passes.
    #[serde(default)]
    keep_polite: bool,
}

#[derive(Serialize)]
struct OptimizeOutput {
    optimized: String,
    original_chars: usize,
    optimized_chars: usize,
    estimated_tokens_before: usize,
    estimated_tokens_after: usize,
    estimated_token_savings_pct: f32,
    rules_applied: Vec<String>,
}

/// OpenAI's rule of thumb for English: ~4 characters per token.
fn estimate_tokens(s: &str) -> usize {
    (s.chars().count() + 3) / 4
}

/// Phrase substitutions: longer verbose form → terse equivalent.
/// Order matters: longer phrases must come before shorter overlapping ones.
fn phrase_substitutions() -> &'static [(&'static str, &'static str)] {
    &[
        ("I would like you to ", ""),
        ("I'd like you to ", ""),
        ("I would like to ", ""),
        ("I'd like to ", ""),
        ("I want you to ", ""),
        ("I need you to ", ""),
        ("Could you please ", ""),
        ("Would you please ", ""),
        ("Can you please ", ""),
        ("Could you kindly ", ""),
        ("Would you kindly ", ""),
        ("Could you ", ""),
        ("Would you ", ""),
        ("Can you ", ""),
        ("Please kindly ", ""),
        ("Please could you ", ""),
        ("Please ", ""),
        ("In order to ", "To "),
        ("in order to ", "to "),
        ("Make sure to ", ""),
        ("make sure to ", ""),
        ("Be sure to ", ""),
        ("be sure to ", ""),
        ("Make sure that ", ""),
        ("make sure that ", ""),
        ("It is important to ", ""),
        ("it is important to ", ""),
        ("It would be great if you could ", ""),
        ("It would be helpful if you could ", ""),
        ("I was wondering if you could ", ""),
        ("Due to the fact that ", "Because "),
        ("due to the fact that ", "because "),
        ("In the event that ", "If "),
        ("in the event that ", "if "),
        ("With regard to ", "About "),
        ("with regard to ", "about "),
        ("With respect to ", "About "),
        ("with respect to ", "about "),
        ("at this point in time ", "now "),
        ("At this point in time ", "Now "),
        ("a large number of ", "many "),
        ("A large number of ", "Many "),
        ("a small number of ", "few "),
        ("A small number of ", "Few "),
        ("Thank you in advance. ", ""),
        ("Thank you in advance ", ""),
        ("Thanks in advance. ", ""),
        ("Thanks in advance ", ""),
        ("Thank you so much. ", ""),
        ("Thank you. ", ""),
        ("Thanks. ", ""),
    ]
}

/// Standalone filler words to delete (whole-word, case-insensitive at word start).
fn filler_words() -> &'static [&'static str] {
    &[
        "very", "really", "quite", "rather", "just", "actually", "basically",
        "literally", "simply", "essentially", "honestly", "frankly", "obviously",
        "clearly", "definitely", "absolutely", "totally", "completely",
    ]
}

fn apply_phrases(input: &str) -> (String, Vec<&'static str>) {
    let mut s = input.to_string();
    let mut applied = Vec::new();
    for (from, to) in phrase_substitutions() {
        if s.contains(from) {
            s = s.replace(from, to);
            applied.push(*from);
        }
    }
    (s, applied)
}

fn strip_fillers(input: &str) -> (String, bool) {
    let mut applied = false;
    let words: Vec<&str> = input.split_whitespace().collect();
    let fillers = filler_words();
    let kept: Vec<&str> = words
        .iter()
        .filter(|w| {
            // Strip surrounding punctuation for comparison.
            let bare = w.trim_matches(|c: char| !c.is_alphanumeric()).to_lowercase();
            if fillers.contains(&bare.as_str()) {
                applied = true;
                false
            } else {
                true
            }
        })
        .copied()
        .collect();
    (kept.join(" "), applied)
}

fn collapse_whitespace(input: &str) -> String {
    let mut out = String::with_capacity(input.len());
    let mut last_was_space = false;
    for c in input.chars() {
        if c.is_whitespace() {
            if !last_was_space {
                out.push(' ');
                last_was_space = true;
            }
        } else {
            out.push(c);
            last_was_space = false;
        }
    }
    out.trim().to_string()
}

#[plugin_fn]
pub fn optimize(input: String) -> FnResult<String> {
    let req: OptimizeInput = serde_json::from_str(&input)
        .map_err(|e| WithReturnCode::new(Error::msg(format!("invalid input: {e}")), 1))?;

    let original = req.prompt.clone();
    let mut working = original.clone();
    let mut rules = Vec::new();

    if !req.keep_polite {
        let (after_phrases, phrases_applied) = apply_phrases(&working);
        working = after_phrases;
        for p in phrases_applied {
            rules.push(format!("removed phrase: \"{}\"", p.trim()));
        }
    }

    let (after_fillers, fillers_applied) = strip_fillers(&working);
    working = after_fillers;
    if fillers_applied {
        rules.push("removed filler words".to_string());
    }

    let collapsed = collapse_whitespace(&working);
    if collapsed.len() != working.trim().len() {
        rules.push("collapsed whitespace".to_string());
    }
    working = collapsed;

    let original_chars = original.chars().count();
    let optimized_chars = working.chars().count();
    let tokens_before = estimate_tokens(&original);
    let tokens_after = estimate_tokens(&working);
    let savings_pct = if tokens_before == 0 {
        0.0
    } else {
        (1.0 - (tokens_after as f32 / tokens_before as f32)) * 100.0
    };

    let out = OptimizeOutput {
        optimized: working,
        original_chars,
        optimized_chars,
        estimated_tokens_before: tokens_before,
        estimated_tokens_after: tokens_after,
        estimated_token_savings_pct: (savings_pct * 10.0).round() / 10.0,
        rules_applied: rules,
    };

    Ok(serde_json::to_string(&out)?)
}
