/**
 * @license
 * Code City Client
 *
 * Copyright 2017 Google Inc.
 * https://codecity.world/
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/**
 * @fileoverview World frame of Code City's client.
 * @author fraser@google.com (Neil Fraser)
 */
'use strict';

var CCC = {};
CCC.World = {};

/**
 * Maximum number of messages saved in history.
 */
CCC.World.maxHistorySize = 10000;

/**
 * Message history.
 */
CCC.World.history = [];

/**
 * Height of history panels.
 * @constant
 */
CCC.World.panelHeight = 256;

/**
 * PID of rate-limiter for resize events.
 */
CCC.World.resizePid = 0;

/**
 * Initialization code called on startup.
 */
CCC.World.init = function() {
  CCC.World.historyDiv = document.getElementById('historyDiv');
  CCC.World.currentDiv = document.getElementById('currentDiv');
  CCC.World.parser = new DOMParser();

  window.addEventListener('resize', CCC.World.resize, false);

  // Report back to the parent frame that we're fully loaded and ready to go.
  parent.postMessage('initWorld', location.origin);
};

/**
 * Receive messages from our parent frame.
 * @param {!Event} e Incoming message event.
 */
CCC.World.receiveMessage = function(e) {
  var origin = e.origin || e.originalEvent.origin;
  if (origin != location.origin) {
    console.error('Message received by world frame from unknown origin: ' +
                  origin);
    return;
  }
  var mode = e.data.mode;
  var text = e.data.text;
  if (mode == 'message') {
    var dom = CCC.World.parser.parseFromString(text, 'text/xml');
    if (dom.getElementsByTagName('parsererror').length) {
      // Not valid XML, treat as string literal.
      console.log(text);
    } else {
      console.log(dom);
    }
  }
};

/**
 * Rerender entire history.
 * Called when the window changes size.
 */
CCC.World.resize = function() {
  clearTimeout(CCC.World.resizePid);
  CCC.World.resizePid = setTimeout(CCC.World.renderHistory, 1000);
};

/**
 * Rerender entire history.
 * Called when the window changes size.
 */
CCC.World.renderHistory = function() {
  var panelBloat = 2 * (5 + 2);  // Margin and border widths must match the CSS.
  CCC.World.historyDiv.innerHTML = '';
  for (var y = 0; y < 3; y++) {
    var rowWidths = CCC.World.rowWidths();
    var rowDiv = document.createElement('div');
    rowDiv.className = 'historyRow';
    for (var x = 0; x < rowWidths.length; x++) {
      var panelDiv = document.createElement('div');
      panelDiv.className = 'historyPanel';
      panelDiv.style.height = (CCC.World.panelHeight) + 'px';
      panelDiv.style.width = (rowWidths[x] - panelBloat) + 'px';
      rowDiv.appendChild(panelDiv);
    }
    CCC.World.historyDiv.appendChild(rowDiv);
  }
};

/**
 * Given the current window width, assign the number and widths of panels on
 * one history row.
 * @return {!Array.<number>} Array of lengths.
 */
CCC.World.rowWidths = function() {
  var windowWidth = CCC.World.historyDiv.offsetWidth;
  var idealWidth = CCC.World.panelHeight * 5 / 4;  // Standard TV ratio.
  var panelCount = Math.round(windowWidth / idealWidth);
  var averageWidth = Math.floor(windowWidth / panelCount);
  averageWidth = Math.max(averageWidth, CCC.World.panelHeight);
  var smallWidth = Math.round(averageWidth * 0.9);
  var largeWidth = averageWidth * 2 - smallWidth;
  // Build an array of lengths.  Add in matching pairs.
  var panels = [];
  for (var i = 0; i < Math.floor(panelCount / 2); i++) {
    if (Math.random() > 0.5) {
      panels.push(averageWidth, averageWidth);
    } else {
      panels.push(smallWidth, largeWidth);
    }
  }
  // Odd number of panels has one in the middle.
  if (panels.length < panelCount) {
    panels.push(averageWidth);
  }
  // Shuffle the array.
  for (var i = panels.length; i; i--) {
    var j = Math.floor(Math.random() * i);
    var temp = panels[i - 1];
    panels[i - 1] = panels[j];
    panels[j] = temp;
  }
  return panels;
};


window.addEventListener('message', CCC.World.receiveMessage, false);
window.addEventListener('load', CCC.World.init, false);
