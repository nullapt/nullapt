//! web-search — query DuckDuckGo's Instant Answer API.
//!
//! Returns the abstract, related topics, and a short list of result links.
//! The only network domain reachable from this skill is api.duckduckgo.com,
//! enforced at the WASM sandbox layer.

use extism_pdk::*;
use serde::{Deserialize, Serialize};

#[derive(Deserialize)]
struct SearchInput {
    query: String,
    /// Optional: cap the number of related-topic results returned. Default 5.
    #[serde(default)]
    limit: Option<usize>,
}

#[derive(Serialize)]
struct SearchResult {
    title: String,
    url: String,
    snippet: String,
}

#[derive(Serialize)]
struct SearchOutput {
    query: String,
    abstract_text: String,
    abstract_source: String,
    abstract_url: String,
    results: Vec<SearchResult>,
}

#[derive(Deserialize)]
struct DDGResponse {
    #[serde(default, rename = "AbstractText")]
    abstract_text: String,
    #[serde(default, rename = "AbstractSource")]
    abstract_source: String,
    #[serde(default, rename = "AbstractURL")]
    abstract_url: String,
    #[serde(default, rename = "RelatedTopics")]
    related_topics: Vec<DDGTopic>,
}

#[derive(Deserialize)]
struct DDGTopic {
    #[serde(default, rename = "Text")]
    text: String,
    #[serde(default, rename = "FirstURL")]
    first_url: String,
}

#[plugin_fn]
pub fn web_search(input: String) -> FnResult<String> {
    let req: SearchInput = serde_json::from_str(&input)
        .map_err(|e| WithReturnCode::new(Error::msg(format!("invalid input: {e}")), 1))?;

    if req.query.trim().is_empty() {
        return Err(WithReturnCode::new(Error::msg("query is required"), 2).into());
    }

    let url = format!(
        "https://api.duckduckgo.com/?q={}&format=json&no_html=1&no_redirect=1",
        urlencode(&req.query)
    );

    let http_req = HttpRequest::new(&url).with_method("GET");
    let resp = http::request::<()>(&http_req, None)
        .map_err(|e| WithReturnCode::new(Error::msg(format!("http request failed: {e}")), 3))?;

    let body_bytes = resp.body();
    let parsed: DDGResponse = serde_json::from_slice(&body_bytes).map_err(|e| {
        WithReturnCode::new(
            Error::msg(format!("parsing duckduckgo response: {e}")),
            4,
        )
    })?;

    let limit = req.limit.unwrap_or(5);
    let results: Vec<SearchResult> = parsed
        .related_topics
        .into_iter()
        .filter(|t| !t.first_url.is_empty() && !t.text.is_empty())
        .take(limit)
        .map(|t| {
            let (title, snippet) = match t.text.split_once(" - ") {
                Some((title, rest)) => (title.to_string(), rest.to_string()),
                None => (t.text.clone(), t.text.clone()),
            };
            SearchResult {
                title,
                url: t.first_url,
                snippet,
            }
        })
        .collect();

    let out = SearchOutput {
        query: req.query,
        abstract_text: parsed.abstract_text,
        abstract_source: parsed.abstract_source,
        abstract_url: parsed.abstract_url,
        results,
    };

    Ok(serde_json::to_string(&out)?)
}

/// Minimal URL-encoder for query strings — handles the characters we expect
/// in a free-text search query without pulling in a full url crate.
fn urlencode(s: &str) -> String {
    let mut out = String::with_capacity(s.len());
    for b in s.bytes() {
        match b {
            b'A'..=b'Z' | b'a'..=b'z' | b'0'..=b'9' | b'-' | b'_' | b'.' | b'~' => {
                out.push(b as char);
            }
            b' ' => out.push('+'),
            _ => out.push_str(&format!("%{:02X}", b)),
        }
    }
    out
}
