

function string_to_json_wrap(v) {
  try {
    if (is_json(v)) {
      return unicode_to_china(JSON.stringify(string_to_json(v), null, '\t'))
    } else {
      return v
    }
  } catch (e) {
  }
  return v
}

function json_wrap_to_string(v) {
  try {
    if (is_json(v)) {
      return unicode_to_china(JSON.stringify(string_to_json(v)))

    } else {
      return v
    }

  } catch (e) {
  }
  return v
}


function string_to_xml_wrap(v) {
  try {
    return vkbeautify.xml(v);
  } catch (e) {
    return v
  }
}

function xml_wrap_to_string(v) {
  try {
    return vkbeautify.xmlmin(v);
  } catch (e) {
    return v
  }
}

function is_json(str) {
  if (typeof str == 'string') {
    try {
      let obj = JSON.parse(str);
      if (typeof obj == 'object' && obj) {
        return true;
      } else {
        return false;
      }

    } catch (e) {
      return false;
    }
  }
}

function check_string_type(v) {
  try {
    // if (v.startsWith("<!DOCTYPE html")) {
    //   return 'HTML'
    // } else if (v.startsWith("<")) {
    if (v.startsWith("<")) {
      return 'XML'
    } else if (is_json(v)) {
      return 'JSON'
    } else {
      return 'TEXT'
    }
  } catch (e) {
    return 'TEXT'
  }
}

function wrap_to_string(v, t) {
  let type = t || check_string_type(v)
  switch (type) {
    case 'JSON': {
      return json_wrap_to_string(v)
    }
    case 'XML': {
      return xml_wrap_to_string(v)
    }
    case 'HTML': {
      return xml_wrap_to_string(v)
    }
  }
  return v
}

function string_to_wrap(v, t) {
  let type = t || check_string_type(v)
  switch (type) {
    case 'JSON': {
      return string_to_json_wrap(v)
    }
    case 'XML': {
      return string_to_xml_wrap(v)
    }
    case 'HTML': {
      return string_to_xml_wrap(v)
    }
  }
  return v
}

function string_to_json(v) {
  try {
    if (is_json(v)) {
      return jsonlint.parse(v)
    } else {
      return v
    }
  } catch (e) {
    return v
  }
}

function unicode_to_china(input) {
  try {
    return input.replace(/\\\\u([0-9a-fA-F]{2,4})/g, function (string, matched) {
      return String.fromCharCode(parseInt(matched, 16))
    })
  } catch (e) {
  }
  return input
}
