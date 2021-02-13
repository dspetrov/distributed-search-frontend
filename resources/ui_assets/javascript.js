$( document ).ready(function() {
    console.log( "ready!" );

    // find elements on the page
    var banner = $("#banner-message");
    var button = $("#submit_button");
    var searchBox = $("#search_text");
    var numResultsBox = $("#num_results");
    var minScoreBox = $("#min_score");
    var resultsTable = $("#results table tbody");
    var resultsWrapper = $("#results");
    var noResultsError = $("#no_results_error");

    // handle search click
    button.on("click", function(){
        banner.addClass("alt");

        // send request to the server
        $.ajax({
          method : "POST",
          contentType: "application/json",
          data: createRequest(),
          url: "documents_search",
          dataType: "json",
          success: onHttpResponse
          });
      });

    function createRequest() {
        var searchQuery = searchBox.val();
        var minScore = parseFloat(minScoreBox.val(), 10);
        if (isNaN(minScore)) {
            minScore = 0;
        }

        var maxNumberOfResults = parseInt(numResultsBox.val());

        if (isNaN(maxNumberOfResults)) {
            maxNumberOfResults = Number.MAX_SAFE_INTEGER;
        }

        // Search request to the server
        var frontEndRequest = {
            SearchQuery: searchQuery,
            MinScore: minScore,
            MaxNumberOfResults: maxNumberOfResults
        };

        return JSON.stringify(frontEndRequest);
    }

    function onHttpResponse(data, status) {
        if (status === "success" ) {
            console.log(data);
            addResults(data);
        } else {
            alert("Error connecting to the server " + status);
        }
    }

    /*
        Add results from the server to the html or how an error message
     */
    function addResults(data) {
        var baseDir = data.DocumentsLocation;

        resultsTable.empty();

        if (data.SearchResults.length == 0) {
            resultsWrapper.hide();
            noResultsError.show();
        } else {
            noResultsError.hide();
            resultsWrapper.show();
        }

        for (var i = 0 ; i < data.SearchResults.length; i++) {
            var title = data.SearchResults[i].Title;
            var extension = data.SearchResults[i].Extension;
            var score = data.SearchResults[i].Score;
            var fullPath = baseDir + "/" + title + "." + extension;
            resultsTable.append("<tr><td><a href=\""+ fullPath + "\">" + title +"</a></td><td>" + score + "</td></tr>");
        }
    }
});
