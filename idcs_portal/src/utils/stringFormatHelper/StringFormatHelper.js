// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

/**
 * Converts the first letter of the sentence to Uppercase and next to lowercase
 * @param {string} sentence Word or sentence to capitalize
 * @returns {string}
 */
export const capitalizeString = (sentence) => {
  if (!sentence.length || sentence.length === 0) {
    return sentence
  }
  if (sentence.length === 1) {
    return sentence.toUpperCase()
  }
  return sentence[0].toUpperCase() + sentence.slice(1)
}
