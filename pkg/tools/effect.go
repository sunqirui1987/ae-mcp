package tools

import (
	"encoding/json"
	"fmt"
	"github.com/sunqirui1987/ae-mcp/pkg/ae"
)

// EffectDetails represents details of an After Effects effect
type EffectDetails map[string]interface{}

// EffectParameters represents parameters for an After Effects effect
type EffectParameters map[string]interface{}

// EffectList represents a list of effects
type EffectList []interface{}

// ApplyEffect applies an effect to a layer in a composition
func ApplyEffect(compName string, layerName string, effectName string, parameters EffectParameters) (EffectDetails, error) {
	// Build the script to apply the effect
	script := `
	try {
		var compName = "` + compName + `";
		var layerName = "` + layerName + `";
		var effectName = "` + effectName + `";
		
		// Find the composition
		var comp = null;
		for (var i = 1; i <= app.project.numItems; i++) {
			var item = app.project.item(i);
			if (item instanceof CompItem && item.name === compName) {
				comp = item;
				break;
			}
		}
		
		if (!comp) {
			return JSON.stringify({
				error: "Composition not found: " + compName
			});
		}
		
		// Find the layer
		var layer = null;
		for (var i = 1; i <= comp.numLayers; i++) {
			var l = comp.layer(i);
			if (l.name === layerName) {
				layer = l;
				break;
			}
		}
		
		if (!layer) {
			return JSON.stringify({
				error: "Layer not found: " + layerName
			});
		}
		
		// Apply the effect
		var effect = layer.Effects.addProperty(effectName);
		if (!effect) {
			return JSON.stringify({
				error: "Failed to apply effect: " + effectName
			});
		}
	`

	// If parameters were provided, adjust them
	if parameters != nil {
		for key, value := range parameters {
			// Different handling based on parameter type
			switch v := value.(type) {
			case float64:
				script += `
				try {
					if (effect.property("` + key + `")) {
						effect.property("` + key + `").setValue(` + fmt.Sprintf("%f", v) + `);
					}
				} catch (paramErr) {
					// Skip this parameter if it doesn't exist or can't be set
				}
				`
			case bool:
				boolVal := 0
				if v {
					boolVal = 1
				}
				script += `
				try {
					if (effect.property("` + key + `")) {
						effect.property("` + key + `").setValue(` + fmt.Sprintf("%d", boolVal) + `);
					}
				} catch (paramErr) {
					// Skip this parameter if it doesn't exist or can't be set
				}
				`
			case string:
				script += `
				try {
					if (effect.property("` + key + `")) {
						effect.property("` + key + `").setValue("` + escapeJSString(v) + `");
					}
				} catch (paramErr) {
					// Skip this parameter if it doesn't exist or can't be set
				}
				`
			case []interface{}:
				// Handle arrays (e.g., for color or point values)
				if len(v) > 0 {
					arrayStr := "["
					for i, item := range v {
						if i > 0 {
							arrayStr += ", "
						}
						switch itemVal := item.(type) {
						case float64:
							arrayStr += fmt.Sprintf("%f", itemVal)
						case int:
							arrayStr += fmt.Sprintf("%d", itemVal)
						case string:
							arrayStr += `"` + escapeJSString(itemVal) + `"`
						default:
							arrayStr += "0"
						}
					}
					arrayStr += "]"
					
					script += `
					try {
						if (effect.property("` + key + `")) {
							effect.property("` + key + `").setValue(` + arrayStr + `);
						}
					} catch (paramErr) {
						// Skip this parameter if it doesn't exist or can't be set
					}
					`
				}
			}
		}
	}

	// Complete the script
	script += `
		// Get information about the applied effect
		var effectInfo = {
			name: effect.name,
			matchName: effect.matchName,
			parameters: []
		};
		
		// Gather parameter information
		for (var i = 1; i <= effect.numProperties; i++) {
			var prop = effect.property(i);
			// Only include properties that we can potentially modify
			if (prop.canSetValue) {
				var paramInfo = {
					name: prop.name,
					matchName: prop.matchName,
					value: prop.value
				};
				effectInfo.parameters.push(paramInfo);
			}
		}
		
		return returnjson(effectInfo);
	} catch (err) {
		return "ERROR: " + err.toString();
	}
	`

	// Execute the script
	result, err := ae.ExecuteScript(script)
	if err != nil {
		return nil, err
	}

	// Extract result
	if resultStr, ok := result.(string); ok {
		// Check if the result indicates an error
		if len(resultStr) > 7 && resultStr[:7] == "ERROR: " {
			return nil, ErrAEScriptError(resultStr[7:])
		}
		
		// Parse the JSON result into a structured object
		var effectDetails EffectDetails
		if err := json.Unmarshal([]byte(resultStr), &effectDetails); err != nil {
			return nil, err
		}
		
		// Check for error in result
		if errMsg, hasErr := effectDetails["error"].(string); hasErr {
			return nil, fmt.Errorf("%s", errMsg)
		}
		
		return effectDetails, nil
	}

	return nil, ErrInvalidResponse
}

// ListAvailableEffects lists all available effects in After Effects
func ListAvailableEffects() (EffectList, error) {
	// Build the script to list all effects
	script := `
	try {
		// Helper function to get effect category by match name
		function getEffectCategory(matchName) {
			// Simple categorization based on common prefixes
			if (matchName.indexOf("ADBE") === 0) {
				if (matchName.indexOf("ADBE Ramp") === 0 || 
					matchName.indexOf("ADBE Laser") === 0 || 
					matchName.indexOf("ADBE Grid") === 0 ||
					matchName.indexOf("ADBE Color") === 0 ||
					matchName.indexOf("ADBE Fill") === 0) {
					return "Generate";
				} else if (matchName.indexOf("ADBE Blur") === 0 ||
						  matchName.indexOf("ADBE Sharpen") === 0) {
					return "Blur & Sharpen";
				} else if (matchName.indexOf("ADBE Channel") === 0) {
					return "Channel";
				} else if (matchName.indexOf("ADBE Noise") === 0 ||
						  matchName.indexOf("ADBE Fractal") === 0) {
					return "Noise & Grain";
				} else if (matchName.indexOf("ADBE Displace") === 0 ||
						  matchName.indexOf("ADBE Distort") === 0 ||
						  matchName.indexOf("ADBE Bulge") === 0 ||
						  matchName.indexOf("ADBE Warp") === 0) {
					return "Distort";
				} else if (matchName.indexOf("ADBE 3D") === 0) {
					return "3D";
				} else if (matchName.indexOf("ADBE Text") === 0) {
					return "Text";
				} else if (matchName.indexOf("ADBE Simulation") === 0) {
					return "Simulation";
				} else {
					return "Utility";
				}
			} else if (matchName.indexOf("AE.ADBE") === 0) {
				return "Expression Controls";
			} else {
				return "Other";
			}
		}

		// Create a dummy composition to test effects
		var tempComp = app.project.items.addComp("Temp Effects Comp", 1920, 1080, 1, 30, 30);
		var tempLayer = tempComp.layers.addSolid([0, 0, 0], "Temp Solid", 1920, 1080, 1);
		
		// Collect effects info
		var effectsInfo = [];
		
		// Get some common categories to test
		var testCategories = [
			"3D Channel",
			"Audio",
			"Blur & Sharpen",
			"Channel",
			"Color Correction",
			"Distort",
			"Expression Controls",
			"Generate",
			"Keying",
			"Matte",
			"Noise & Grain",
			"Perspective",
			"Simulation",
			"Stylize",
			"Text",
			"Time",
			"Transition",
			"Utility"
		];
		
		// First try common effects with known match names
		var commonEffects = [
			{ name: "Gaussian Blur", matchName: "ADBE Gaussian Blur 2" },
			{ name: "Fast Blur", matchName: "ADBE Fast Blur" },
			{ name: "Color Correction", matchName: "ADBE Color Balance" },
			{ name: "Hue/Saturation", matchName: "ADBE Easy Levels" },
			{ name: "Glow", matchName: "ADBE Glo2" },
			{ name: "Noise", matchName: "ADBE Noise" },
			{ name: "Fill", matchName: "ADBE Fill" },
			{ name: "Drop Shadow", matchName: "ADBE Drop Shadow" },
			{ name: "Exposure", matchName: "ADBE Exposure" },
			{ name: "Directional Blur", matchName: "ADBE Motion Blur" },
			{ name: "Turbulent Displace", matchName: "ADBE Turbulent Displace" },
			{ name: "Fractal Noise", matchName: "ADBE Fractal Noise" },
			{ name: "Transform", matchName: "ADBE Geometry2" },
			{ name: "Curves", matchName: "ADBE CurvesCustom" },
			{ name: "Levels", matchName: "ADBE Easy Levels2" }
		];

		for (var i = 0; i < commonEffects.length; i++) {
			try {
				var effect = tempLayer.Effects.addProperty(commonEffects[i].matchName);
				
				if (effect) {
					var effectInfo = {
						name: effect.name,
						matchName: effect.matchName,
						displayName: commonEffects[i].name,
						category: getEffectCategory(effect.matchName),
						parameters: []
					};
					
					// Get parameters
					for (var p = 1; p <= effect.numProperties; p++) {
						var prop = effect.property(p);
						if (prop.canSetValue) {
							effectInfo.parameters.push({
								name: prop.name,
								matchName: prop.matchName
							});
						}
					}
					
					effectsInfo.push(effectInfo);
					
					// Remove the effect
					effect.remove();
				}
			} catch (e) {
				// Skip this effect if it can't be added
			}
		}
		
		// Remove the temp composition
		tempComp.remove();
		
		return returnjson(effectsInfo);
	} catch (err) {
		return "ERROR: " + err.toString();
	}
	`

	// Execute the script
	result, err := ae.ExecuteScript(script)
	if err != nil {
		return nil, err
	}

	// Extract result
	if resultStr, ok := result.(string); ok {
		// Check if the result indicates an error
		if len(resultStr) > 7 && resultStr[:7] == "ERROR: " {
			return nil, ErrAEScriptError(resultStr[7:])
		}
		
		// Parse the JSON result into a structured object
		var effectsList EffectList
		if err := json.Unmarshal([]byte(resultStr), &effectsList); err != nil {
			return nil, err
		}
		
		return effectsList, nil
	}

	return nil, ErrInvalidResponse
}

// GetEffectParameters gets all parameters of an effect applied to a layer
func GetEffectParameters(compName string, layerName string, effectName string) (EffectDetails, error) {
	// Build the script to get effect parameters
	script := `
	try {
		var compName = "` + compName + `";
		var layerName = "` + layerName + `";
		var effectName = "` + effectName + `";
		
		// Find the composition
		var comp = null;
		for (var i = 1; i <= app.project.numItems; i++) {
			var item = app.project.item(i);
			if (item instanceof CompItem && item.name === compName) {
				comp = item;
				break;
			}
		}
		
		if (!comp) {
			return JSON.stringify({
				error: "Composition not found: " + compName
			});
		}
		
		// Find the layer
		var layer = null;
		for (var i = 1; i <= comp.numLayers; i++) {
			var l = comp.layer(i);
			if (l.name === layerName) {
				layer = l;
				break;
			}
		}
		
		if (!layer) {
			return JSON.stringify({
				error: "Layer not found: " + layerName
			});
		}
		
		// Find the effect
		var effect = null;
		for (var i = 1; i <= layer.Effects.numProperties; i++) {
			var e = layer.Effects.property(i);
			if (e.name === effectName) {
				effect = e;
				break;
			}
		}
		
		if (!effect) {
			return JSON.stringify({
				error: "Effect not found: " + effectName
			});
		}
		
		// Get effect parameters
		var effectInfo = {
			name: effect.name,
			matchName: effect.matchName,
			parameters: []
		};
		
		for (var i = 1; i <= effect.numProperties; i++) {
			var prop = effect.property(i);
			
			if (prop.canSetValue) {
				var paramInfo = {
					name: prop.name,
					matchName: prop.matchName,
					value: prop.value
				};
				
				// Get information about possible values if it's a dropdown menu
				if (prop.propertyValueType === PropertyValueType.OneD && prop.hasMax) {
					paramInfo.min = prop.minValue;
					paramInfo.max = prop.maxValue;
				}
				
				effectInfo.parameters.push(paramInfo);
			}
		}
		
		return returnjson(effectInfo);
	} catch (err) {
		return "ERROR: " + err.toString();
	}
	`

	// Execute the script
	result, err := ae.ExecuteScript(script)
	if err != nil {
		return nil, err
	}

	// Extract result
	if resultStr, ok := result.(string); ok {
		// Check if the result indicates an error
		if len(resultStr) > 7 && resultStr[:7] == "ERROR: " {
			return nil, ErrAEScriptError(resultStr[7:])
		}
		
		// Parse the JSON result into a structured object
		var effectDetails EffectDetails
		if err := json.Unmarshal([]byte(resultStr), &effectDetails); err != nil {
			return nil, err
		}
		
		// Check for error in result
		if errMsg, hasErr := effectDetails["error"].(string); hasErr {
			return nil, fmt.Errorf("%s", errMsg)
		}
		
		return effectDetails, nil
	}

	return nil, ErrInvalidResponse
}

// ModifyEffectParameters modifies parameters of an existing effect on a layer
func ModifyEffectParameters(compName string, layerName string, effectName string, parameters EffectParameters) (EffectDetails, error) {
	// Build the script to modify effect parameters
	script := `
	try {
		var compName = "` + compName + `";
		var layerName = "` + layerName + `";
		var effectName = "` + effectName + `";
		
		// Find the composition
		var comp = null;
		for (var i = 1; i <= app.project.numItems; i++) {
			var item = app.project.item(i);
			if (item instanceof CompItem && item.name === compName) {
				comp = item;
				break;
			}
		}
		
		if (!comp) {
			return JSON.stringify({
				error: "Composition not found: " + compName
			});
		}
		
		// Find the layer
		var layer = null;
		for (var i = 1; i <= comp.numLayers; i++) {
			var l = comp.layer(i);
			if (l.name === layerName) {
				layer = l;
				break;
			}
		}
		
		if (!layer) {
			return JSON.stringify({
				error: "Layer not found: " + layerName
			});
		}
		
		// Find the effect
		var effect = null;
		for (var i = 1; i <= layer.Effects.numProperties; i++) {
			var e = layer.Effects.property(i);
			if (e.name === effectName) {
				effect = e;
				break;
			}
		}
		
		if (!effect) {
			return JSON.stringify({
				error: "Effect not found: " + effectName
			});
		}
		
		// Keep track of changes
		var modifiedParameters = [];
	`

	// Add parameter modifications to the script
	for key, value := range parameters {
		// Different handling based on parameter type
		switch v := value.(type) {
		case float64:
			script += `
			try {
				// Find parameter by name
				for (var i = 1; i <= effect.numProperties; i++) {
					var prop = effect.property(i);
					if (prop.name === "` + key + `" && prop.canSetValue) {
						prop.setValue(` + fmt.Sprintf("%f", v) + `);
						modifiedParameters.push({ name: "` + key + `", value: ` + fmt.Sprintf("%f", v) + ` });
						break;
					}
				}
			} catch (paramErr) {
				// Skip this parameter if it can't be set
			}
			`
		case bool:
			boolVal := 0
			if v {
				boolVal = 1
			}
			script += `
			try {
				// Find parameter by name
				for (var i = 1; i <= effect.numProperties; i++) {
					var prop = effect.property(i);
					if (prop.name === "` + key + `" && prop.canSetValue) {
						prop.setValue(` + fmt.Sprintf("%d", boolVal) + `);
						modifiedParameters.push({ name: "` + key + `", value: ` + fmt.Sprintf("%v", v) + ` });
						break;
					}
				}
			} catch (paramErr) {
				// Skip this parameter if it can't be set
			}
			`
		case string:
			script += `
			try {
				// Find parameter by name
				for (var i = 1; i <= effect.numProperties; i++) {
					var prop = effect.property(i);
					if (prop.name === "` + key + `" && prop.canSetValue) {
						prop.setValue("` + escapeJSString(v) + `");
						modifiedParameters.push({ name: "` + key + `", value: "` + escapeJSString(v) + `" });
						break;
					}
				}
			} catch (paramErr) {
				// Skip this parameter if it can't be set
			}
			`
		case []interface{}:
			// Handle arrays (e.g., for color or point values)
			if len(v) > 0 {
				arrayStr := "["
				for i, item := range v {
					if i > 0 {
						arrayStr += ", "
					}
					switch itemVal := item.(type) {
					case float64:
						arrayStr += fmt.Sprintf("%f", itemVal)
					case int:
						arrayStr += fmt.Sprintf("%d", itemVal)
					case string:
						arrayStr += `"` + escapeJSString(itemVal) + `"`
					default:
						arrayStr += "0"
					}
				}
				arrayStr += "]"
				
				script += `
				try {
					// Find parameter by name
					for (var i = 1; i <= effect.numProperties; i++) {
						var prop = effect.property(i);
						if (prop.name === "` + key + `" && prop.canSetValue) {
							prop.setValue(` + arrayStr + `);
							modifiedParameters.push({ name: "` + key + `", value: ` + arrayStr + ` });
							break;
						}
					}
				} catch (paramErr) {
					// Skip this parameter if it can't be set
				}
				`
			}
		}
	}

	// Complete the script
	script += `
		// Get updated effect information
		var effectInfo = {
			name: effect.name,
			matchName: effect.matchName,
			modifiedParameters: modifiedParameters,
			currentParameters: []
		};
		
		for (var i = 1; i <= effect.numProperties; i++) {
			var prop = effect.property(i);
			if (prop.canSetValue) {
				effectInfo.currentParameters.push({
					name: prop.name,
					value: prop.value
				});
			}
		}
		
		return returnjson(effectInfo);
	} catch (err) {
		return "ERROR: " + err.toString();
	}
	`

	// Execute the script
	result, err := ae.ExecuteScript(script)
	if err != nil {
		return nil, err
	}

	// Extract result
	if resultStr, ok := result.(string); ok {
		// Check if the result indicates an error
		if len(resultStr) > 7 && resultStr[:7] == "ERROR: " {
			return nil, ErrAEScriptError(resultStr[7:])
		}
		
		// Parse the JSON result into a structured object
		var effectDetails EffectDetails
		if err := json.Unmarshal([]byte(resultStr), &effectDetails); err != nil {
			return nil, err
		}
		
		// Check for error in result
		if errMsg, hasErr := effectDetails["error"].(string); hasErr {
			return nil, fmt.Errorf("%s", errMsg)
		}
		
		return effectDetails, nil
	}

	return nil, ErrInvalidResponse
} 