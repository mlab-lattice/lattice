# Lattice API doc generation

To generate lattice api docs, you need to install the following tools first:

1- https://github.com/swaggo/swag: Converts Go annotations to swagger docs.
 
``$ go get -u github.com/swaggo/swag/cmd/swag``

2- https://github.com/apiaryio/swagger2blueprint: Converts swagger docs to API blueprint docs.

``$ npm install -g swagger2blueprint``

3- https://github.com/danielgtaylor/aglio: Renders API Blueprint

``$ npm install -g aglio``  

After you have all these installed, run:


``$ make api-docs``

This will generate `lattice-api.html` under lattice/pkg/api/docs