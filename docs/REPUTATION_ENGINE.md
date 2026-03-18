# TrueConnect Reputation Engine

## Overview
The Reputation Engine computes a real-time **Trust Score** (0-100) for each user based on their interactions, weighted by the credibility of their peers. This ensures that a Sybil attack (creating many fake accounts to boost a score) is mathematically ineffective, as fake accounts lack high trust weights and identity verification.

## Core Formula
The engine utilizes a **Weighted Bayesian Average** calculated via Neo4j Graph queries. 

### 1. Peer Weighting
Each rating ($r$) received from another user ($rater$) is scaled by two factors:
1. **Identity Weight ($w_{id}$)**: Raters with `id_verified` or `photo_verified` KYC status get a `1.5x` multiplier. Unverified users get `1.0x`.
2. **Trust Weight ($w_{trust}$)**: The rater's own Trust Score is normalized as a coefficient ($TrustScore / 100.0$). A rater with an 80 score has twice the influence of a rater with a 40 score.

$$ Rating_{effective} = r.score \times w_{id} \times w_{trust} $$

### 2. Bayesian Smoothing
To prevent new users with just one 5-star rating from immediately jumping to a 100 trust score, we apply Bayesian smoothing toward a neutral baseline (equivalent to a score of 2.5 out of 5.0, weighted by an arbitrary factor of $C=5$ ratings).

$$ Smoothed = \frac{(\text{Average}_{effective} \times \text{Count}) + (2.5 \times 5)}{\text{Count} + 5} $$

### 3. Final Conversion
The final score is multiplied by 20 to map the 0-5 scale onto a **0-100 scale**, returning an integer:

$$ TrustScore = Smoothed \times 20 $$

## Trust Engine Worker (`internal/worker/trust_engine.go`)
- Ratings are processed asynchronously to avoid blocking user interactions.
- When an interaction is explicitly verified (`ConfirmInteraction` in Neo4j), an event is dispatched to the background `TrustEngine`.
- The engine computes the new score directly within Neo4j using Cypher.
- The new score is updated across Neo4j, PostgreSQL (for quick SQL joins), and Redis (for sub-millisecond cache retrieval in the matchmaking swiping feed).
