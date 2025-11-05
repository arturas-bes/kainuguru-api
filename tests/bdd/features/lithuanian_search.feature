Feature: Lithuanian Language Search Support
  As a Lithuanian user
  I want to search for products using Lithuanian language with proper diacritics handling
  So that I can find products regardless of how I type Lithuanian characters

  Background:
    Given the system has Lithuanian FTS configuration enabled
    And there are products with Lithuanian names containing diacritics
    And the search indexes are optimized for Lithuanian language

  Scenario: Search with Lithuanian diacritics
    Given there are products named "ąčęėįšųūž duona", "pienas", "mėsa"
    When I search for "ąčęėįšųūž" via GraphQL
    Then I should see products containing "ąčęėįšųūž duona"
    And the search should handle diacritics correctly
    And results should be ranked by relevance

  Scenario: Search without diacritics matches Lithuanian characters
    Given there are products named "ąčęėįšųūž duona", "sūris", "žuvis"
    When I search for "aceeisuuz" (without diacritics) via GraphQL
    Then I should see products containing "ąčęėįšųūž"
    And the search should normalize Lithuanian characters
    And "aceeisuuz" should match "ąčęėįšųūž"

  Scenario: Bidirectional diacritics matching
    Given there are products with both "pienas" and "pieñas" (with foreign diacritics)
    When I search for "pienas" via GraphQL
    Then I should see both products
    And the search should handle various diacritic forms
    And Lithuanian "ė" should match foreign "ē" and "ê"

  Scenario: Case-insensitive Lithuanian search
    Given there are products named "PIENAS", "Pienas", "pienas"
    When I search for "pienas" via GraphQL
    Then I should see all three products
    And case should not affect search results
    And "ĄČĘĖĮŠŲŪŽ" should match "ąčęėįšųūž"

  Scenario: Lithuanian word stemming
    Given there are products named "duonos", "duoną", "duona"
    When I search for "duona" via GraphQL
    Then I should see all variations of "duona"
    And Lithuanian word forms should be matched
    And grammatical endings should be handled

  Scenario: Lithuanian stop words filtering
    Given there are products named "ir pienas", "su duona", "be cukraus"
    When I search for "ir pienas" via GraphQL
    Then Lithuanian stop words "ir", "su", "be" should be filtered
    And the search should focus on meaningful words
    And results should prioritize "pienas" over "ir"

  Scenario: Lithuanian text normalization in product names
    Given there are products with mixed encoding and diacritics
    When I search for normalized Lithuanian terms via GraphQL
    Then all products should be found regardless of original encoding
    And text should be consistently normalized
    And UTF-8 encoding should be properly handled

  Scenario: Lithuanian compound word search
    Given there are products named "šaldytas pienas", "rūkytos dešros"
    When I search for "šaldytas" via GraphQL
    Then I should see "šaldytas pienas"
    And compound words should be searchable by parts
    And "rūkytos" should find "rūkytos dešros"

  Scenario: Lithuanian abbreviations and brand names
    Given there are products with Lithuanian brand names "AB Rokiškio sūris"
    When I search for "Rokiškio" via GraphQL
    Then I should see products from that brand
    And abbreviations like "AB" should be handled appropriately
    And brand name search should work with diacritics

  Scenario: Mixed Lithuanian and English search
    Given there are products named "Coca Cola", "lietuviškas pienas"
    When I search for "lietuviškas cola" via GraphQL
    Then I should see relevant products for each term
    And mixed language search should be supported
    And results should handle both languages appropriately

  Scenario: Lithuanian phonetic similarity
    Given there are products with similar sounding Lithuanian names
    When I search for phonetically similar terms via GraphQL
    Then I should see products with similar pronunciation
    And phonetic matching should complement exact matching
    And "š" should be somewhat similar to "s"

  Scenario: Lithuanian seasonal and regional terms
    Given there are products with seasonal terms like "žieminis", "vasarinis"
    When I search for these seasonal terms via GraphQL
    Then I should see appropriate seasonal products
    And regional Lithuanian terms should be recognized
    And context-specific vocabulary should be handled

  Scenario: Performance with Lithuanian text processing
    Given there are thousands of products with Lithuanian names
    When I search for common Lithuanian terms via GraphQL
    Then the search should complete within 200ms
    And Lithuanian text processing should not degrade performance
    And response times should be consistent for all character types

  Scenario: Lithuanian search suggestions
    Given users frequently search for Lithuanian terms
    When I request search suggestions for partial Lithuanian words via GraphQL
    Then I should see relevant Lithuanian completions
    And suggestions should handle diacritics properly
    And popular Lithuanian grocery terms should be prioritized

  Scenario: Error handling with invalid Lithuanian input
    When I search with malformed Lithuanian characters via GraphQL
    Then the search should handle invalid input gracefully
    And error messages should be appropriate
    And the system should not crash on malformed UTF-8

  Scenario: Lithuanian full-text search ranking
    Given there are products with Lithuanian descriptions containing target terms
    When I search for Lithuanian terms via GraphQL
    Then results should be ranked by Lithuanian FTS relevance
    And name matches should rank higher than description matches
    And exact diacritic matches should rank higher than normalized matches