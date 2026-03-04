module.exports = function (eleventyConfig) {
  eleventyConfig.addPassthroughCopy({ "src/assets": "assets" });
  eleventyConfig.addPassthroughCopy({ "website/assets/media": "assets/media" });
  eleventyConfig.addPassthroughCopy({ "website/about/woosunho": "about/woosunho" });
  eleventyConfig.addPassthroughCopy({ "website/lab": "lab" });
  eleventyConfig.addPassthroughCopy({ "website/privacy.html": "privacy.html" });
  eleventyConfig.addPassthroughCopy({ "website/terms.html": "terms.html" });
  eleventyConfig.addPassthroughCopy({ "website/payment-success.html": "payment-success.html" });
  eleventyConfig.addPassthroughCopy({ "website/payment-fail.html": "payment-fail.html" });

  return {
    dir: {
      input: "src",
      includes: "_includes",
      data: "_data",
      output: "website"
    },
    htmlTemplateEngine: "njk",
    markdownTemplateEngine: "njk"
  };
};
