/**
 * Pydio Cells Rest API
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * OpenAPI spec version: 1.0
 * 
 *
 * NOTE: This class is auto generated by the swagger code generator program.
 * https://github.com/swagger-api/swagger-codegen.git
 * Do not edit the class manually.
 *
 */

'use strict';

exports.__esModule = true;

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError('Cannot call a class as a function'); } }

var _ApiClient = require('../ApiClient');

var _ApiClient2 = _interopRequireDefault(_ApiClient);

/**
* The RestDeleteVirtualNodeResponse model module.
* @module model/RestDeleteVirtualNodeResponse
* @version 1.0
*/

var RestDeleteVirtualNodeResponse = (function () {
    /**
    * Constructs a new <code>RestDeleteVirtualNodeResponse</code>.
    * @alias module:model/RestDeleteVirtualNodeResponse
    * @class
    */

    function RestDeleteVirtualNodeResponse() {
        _classCallCheck(this, RestDeleteVirtualNodeResponse);

        this.Success = undefined;
    }

    /**
    * Constructs a <code>RestDeleteVirtualNodeResponse</code> from a plain JavaScript object, optionally creating a new instance.
    * Copies all relevant properties from <code>data</code> to <code>obj</code> if supplied or a new instance if not.
    * @param {Object} data The plain JavaScript object bearing properties of interest.
    * @param {module:model/RestDeleteVirtualNodeResponse} obj Optional instance to populate.
    * @return {module:model/RestDeleteVirtualNodeResponse} The populated <code>RestDeleteVirtualNodeResponse</code> instance.
    */

    RestDeleteVirtualNodeResponse.constructFromObject = function constructFromObject(data, obj) {
        if (data) {
            obj = obj || new RestDeleteVirtualNodeResponse();

            if (data.hasOwnProperty('Success')) {
                obj['Success'] = _ApiClient2['default'].convertToType(data['Success'], 'Boolean');
            }
        }
        return obj;
    };

    /**
    * @member {Boolean} Success
    */
    return RestDeleteVirtualNodeResponse;
})();

exports['default'] = RestDeleteVirtualNodeResponse;
module.exports = exports['default'];